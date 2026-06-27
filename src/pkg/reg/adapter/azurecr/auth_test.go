// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package azurecr

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/pkg/reg/model"
)

var (
	mockURL      = "https://test.azurecr.io"
	mockUsername = "user"
	mockPassword = "password"
	mockToken    = "test-token"
)

func TestAuth(t *testing.T) {
	// mock v2 API
	httpmock.RegisterResponder(http.MethodGet, mockURL+"/v2/",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(http.StatusUnauthorized, "")
			resp.Header.Set("Www-Authenticate", `Bearer realm="https://test.azurecr.io/oauth2/token",service="test.azurecr.io"`)
			return resp, nil
		})
	// mock token API
	httpmock.RegisterResponderWithQuery(http.MethodGet, mockURL+"/oauth2/token",
		url.Values{
			"service": []string{"test.azurecr.io"},
			"scope":   []string{"repository:library/busybox:metadata_read"},
		},
		func(req *http.Request) (*http.Response, error) {
			username, password, ok := req.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, mockUsername, username)
			assert.Equal(t, mockPassword, password)
			return httpmock.NewStringResponse(http.StatusOK, fmt.Sprintf(`{"access_token": "%s"}`, mockToken)), nil
		})

	a := newAuthorizer(&model.Registry{URL: mockURL, Credential: &model.Credential{AccessKey: mockUsername, AccessSecret: mockPassword}})
	ct := &http.Client{}
	a.client = ct
	httpmock.ActivateNonDefault(ct)
	t.Cleanup(httpmock.DeactivateAndReset)

	// test authorize
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", mockURL, "/v2/library/busybox/tags/list"), nil)
	assert.NoError(t, err)
	err = a.Modify(req)
	assert.NoError(t, err)
	// check whether set bearer token
	tokenHeader := req.Header.Get("Authorization")
	assert.Equal(t, fmt.Sprintf("Bearer %s", mockToken), tokenHeader)
}
