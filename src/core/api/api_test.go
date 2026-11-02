// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/member"
	memberModels "github.com/goharbor/harbor/src/pkg/member/models"
	"github.com/goharbor/harbor/src/pkg/user"
)

var (
	nonSysAdminID, projAdminID, projDeveloperID, projGuestID, projLimitedGuestID, projAdminRobotID int64
	projAdminPMID, projDeveloperPMID, projGuestPMID, projLimitedGuestPMID, projAdminRobotPMID      int
	// The following users/credentials are registered and assigned roles at the beginning of
	// running testing and cleaned up at the end.
	// Do not try to change the system and project roles that the users have during
	// the testing. Creating a new one in your own case if needed.
	// The project roles that the users have are for project library.
	sysAdmin = &usrInfo{
		Name:   "admin",
		Passwd: "Harbor12345",
	}
	nonSysAdmin = &usrInfo{
		Name:   "non_admin",
		Passwd: "Harbor12345",
	}
	projAdmin = &usrInfo{
		Name:   "proj_admin",
		Passwd: "Harbor12345",
	}
	projDeveloper = &usrInfo{
		Name:   "proj_developer",
		Passwd: "Harbor12345",
	}
	projGuest = &usrInfo{
		Name:   "proj_guest",
		Passwd: "Harbor12345",
	}
)

type testingRequest struct {
	method      string
	url         string
	header      http.Header
	queryStruct any
	bodyJSON    any
	credential  *usrInfo
}

type codeCheckingCase struct {
	request  *testingRequest
	code     int
	postFunc func(*httptest.ResponseRecorder) error
}

type testRequestBuilder struct {
	method string
	rawURL string
	header http.Header
	body   io.Reader
}

func newTestRequestBuilder() *testRequestBuilder {
	return &testRequestBuilder{header: http.Header{}}
}

func (b *testRequestBuilder) Get(rawURL string) *testRequestBuilder {
	return b.withMethod(http.MethodGet, rawURL)
}

func (b *testRequestBuilder) Post(rawURL string) *testRequestBuilder {
	return b.withMethod(http.MethodPost, rawURL)
}

func (b *testRequestBuilder) Put(rawURL string) *testRequestBuilder {
	return b.withMethod(http.MethodPut, rawURL)
}

func (b *testRequestBuilder) Delete(rawURL string) *testRequestBuilder {
	return b.withMethod(http.MethodDelete, rawURL)
}

func (b *testRequestBuilder) Head(rawURL string) *testRequestBuilder {
	return b.withMethod(http.MethodHead, rawURL)
}

func (b *testRequestBuilder) Patch(rawURL string) *testRequestBuilder {
	return b.withMethod(http.MethodPatch, rawURL)
}

func (b *testRequestBuilder) withMethod(method, rawURL string) *testRequestBuilder {
	b.method = method
	b.rawURL = rawURL
	return b
}

func (b *testRequestBuilder) Add(key, value string) *testRequestBuilder {
	b.header.Add(key, value)
	return b
}

func (b *testRequestBuilder) Set(key, value string) *testRequestBuilder {
	b.header.Set(key, value)
	return b
}

func (b *testRequestBuilder) SetBasicAuth(username, password string) *testRequestBuilder {
	authReq, _ := http.NewRequest(http.MethodGet, "/", nil)
	authReq.SetBasicAuth(username, password)
	b.header.Set("Authorization", authReq.Header.Get("Authorization"))
	return b
}

func (b *testRequestBuilder) QueryStruct(queryStruct any) (*testRequestBuilder, error) {
	values, err := queryValues(queryStruct)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return b, nil
	}
	sep := "?"
	if strings.Contains(b.rawURL, "?") {
		sep = "&"
	}
	b.rawURL += sep + values.Encode()
	return b, nil
}

func (b *testRequestBuilder) BodyJSON(body any) (*testRequestBuilder, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	b.body = strings.NewReader(string(payload))
	b.header.Set("Content-Type", "application/json")
	return b, nil
}

func (b *testRequestBuilder) Request() (*http.Request, error) {
	req, err := http.NewRequest(b.method, b.rawURL, b.body)
	if err != nil {
		return nil, err
	}
	req.Header = b.header.Clone()
	return req, nil
}

func queryValues(queryStruct any) (url.Values, error) {
	values := url.Values{}
	rv := reflect.Indirect(reflect.ValueOf(queryStruct))
	if !rv.IsValid() {
		return values, nil
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("queryStruct must be a struct")
	}

	rt := rv.Type()
	for i := range rv.NumField() {
		field := rt.Field(i)
		name := strings.Split(field.Tag.Get("url"), ",")[0]
		if name == "" || name == "-" {
			name = field.Name
		}
		value := rv.Field(i)
		if value.IsZero() {
			continue
		}
		values.Set(name, fmt.Sprint(value.Interface()))
	}
	return values, nil
}

func newRequest(r *testingRequest) (*http.Request, error) {
	if r == nil {
		return nil, nil
	}

	reqBuilder := newTestRequestBuilder()
	switch strings.ToUpper(r.method) {
	case "", http.MethodGet:
		reqBuilder = reqBuilder.Get(r.url)
	case http.MethodPost:
		reqBuilder = reqBuilder.Post(r.url)
	case http.MethodPut:
		reqBuilder = reqBuilder.Put(r.url)
	case http.MethodDelete:
		reqBuilder = reqBuilder.Delete(r.url)
	case http.MethodHead:
		reqBuilder = reqBuilder.Head(r.url)
	case http.MethodPatch:
		reqBuilder = reqBuilder.Patch(r.url)
	default:
		return nil, fmt.Errorf("unsupported method %s", r.method)
	}

	for key, values := range r.header {
		for _, value := range values {
			reqBuilder = reqBuilder.Add(key, value)
		}
	}

	if r.queryStruct != nil {
		var err error
		reqBuilder, err = reqBuilder.QueryStruct(r.queryStruct)
		if err != nil {
			return nil, err
		}
	}

	if r.bodyJSON != nil {
		var err error
		reqBuilder, err = reqBuilder.BodyJSON(r.bodyJSON)
		if err != nil {
			return nil, err
		}
	}

	if r.credential != nil {
		reqBuilder = reqBuilder.SetBasicAuth(r.credential.Name, r.credential.Passwd)
	}

	return reqBuilder.Request()
}

func handle(r *testingRequest) (*httptest.ResponseRecorder, error) {
	req, err := newRequest(r)
	if err != nil {
		return nil, err
	}

	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	return resp, nil
}

func handleAndParse(r *testingRequest, v any) error {
	resp, err := handle(r)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.Code >= 200 && resp.Code <= 299 {
		return json.Unmarshal(data, v)
	}

	return &common_http.Error{
		Code:    resp.Code,
		Message: string(data),
	}
}

func runCodeCheckingCases(t *testing.T, cases ...*codeCheckingCase) {
	for i, c := range cases {
		t.Logf("running case %d ...", i)
		resp, err := handle(c.request)
		require.Nil(t, err)
		equal := assert.Equal(t, c.code, resp.Code)
		if !equal {
			if resp.Body.Len() > 0 {
				t.Log(resp.Body.String())
			}
			t.FailNow()
		}

		if c.postFunc != nil {
			if err := c.postFunc(resp); err != nil {
				t.Logf("error in running post function: %v", err)
				t.Error(err)
			}
		}
	}
}

func parseResourceID(resp *httptest.ResponseRecorder) (int64, error) {
	location := resp.Header().Get(http.CanonicalHeaderKey("location"))
	if len(location) == 0 {
		return 0, fmt.Errorf("empty location header")
	}
	index := strings.LastIndex(location, "/")
	if index == -1 {
		return 0, fmt.Errorf("location header %s contains no /", location)
	}

	id := strings.TrimPrefix(location, location[:index+1])
	if len(id) == 0 {
		return 0, fmt.Errorf("location header %s contains no resource ID", location)
	}

	return strconv.ParseInt(id, 10, 64)
}

func TestMain(m *testing.M) {
	if err := prepare(); err != nil {
		panic(err)
	}
	dao.ExecuteBatchSQL([]string{
		"insert into user_group (group_name, group_type, ldap_group_dn) values ('test_group_01_api', 1, 'cn=harbor_users,ou=sample,ou=vmware,dc=harbor,dc=com')",
		"insert into user_group (group_name, group_type, ldap_group_dn) values ('vsphere.local\\administrators', 2, '')",
	})

	defer dao.ExecuteBatchSQL([]string{
		"delete from harbor_label",
		"delete from robot",
		"delete from user_group",
		"delete from project_member where id > 1",
	})

	ret := m.Run()
	clean()
	os.Exit(ret)
}

func prepare() error {
	ctx := orm.Context()
	// register nonSysAdmin
	var err error
	nsID, err := user.Mgr.Create(ctx, &models.User{
		Username: nonSysAdmin.Name,
		Password: nonSysAdmin.Passwd,
		Email:    nonSysAdmin.Name + "@test.com",
	})
	if err != nil {
		return err
	}
	nonSysAdminID = int64(nsID)

	// register projAdmin and assign project admin role

	paID, err := user.Mgr.Create(ctx, &models.User{
		Username: projAdmin.Name,
		Password: projAdmin.Passwd,
		Email:    projAdmin.Name + "@test.com",
	})
	if err != nil {
		return err
	}
	projAdminID = int64(paID)
	if projAdminPMID, err = member.Mgr.AddProjectMember(ctx, memberModels.Member{
		ProjectID:  1,
		Role:       common.RoleProjectAdmin,
		EntityID:   int(projAdminID),
		EntityType: common.UserMember,
	}); err != nil {
		return err
	}

	// register projDeveloper and assign project developer role
	pdID, err := user.Mgr.Create(ctx, &models.User{
		Username: projDeveloper.Name,
		Password: projDeveloper.Passwd,
		Email:    projDeveloper.Name + "@test.com",
	})
	if err != nil {
		return err
	}
	projDeveloperID = int64(pdID)

	if projDeveloperPMID, err = member.Mgr.AddProjectMember(ctx, memberModels.Member{
		ProjectID:  1,
		Role:       common.RoleDeveloper,
		EntityID:   int(projDeveloperID),
		EntityType: common.UserMember,
	}); err != nil {
		return err
	}

	// register projGuest and assign project guest role
	pgID, err := user.Mgr.Create(ctx, &models.User{
		Username: projGuest.Name,
		Password: projGuest.Passwd,
		Email:    projGuest.Name + "@test.com",
	})
	if err != nil {
		return err
	}
	projGuestID = int64(pgID)

	if projGuestPMID, err = member.Mgr.AddProjectMember(ctx, memberModels.Member{
		ProjectID:  1,
		Role:       common.RoleGuest,
		EntityID:   int(projGuestID),
		EntityType: common.UserMember,
	}); err != nil {
		return err
	}
	return err
}

func clean() {
	ctx := orm.Context()
	pmids := []int{projAdminPMID, projDeveloperPMID, projGuestPMID}

	for _, id := range pmids {
		if err := member.Mgr.Delete(ctx, 1, id); err != nil {
			fmt.Printf("failed to clean up member %d from project library: %v", id, err)
		}
	}
	userids := []int64{nonSysAdminID, projAdminID, projDeveloperID, projGuestID}
	for _, id := range userids {
		if err := user.Mgr.Delete(ctx, int(id)); err != nil {
			fmt.Printf("failed to clean up user %d: %v \n", id, err)
		}
	}
}
