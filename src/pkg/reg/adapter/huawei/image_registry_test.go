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

package huawei

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/docker/distribution"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/pkg/reg/model"
)

const mockHuaweiURL = "https://swr.cn-north-1.myhuaweicloud.com"

func registerHuaweiResponder(method, path string, status int, body string) {
	httpmock.RegisterResponder(method, mockHuaweiURL+path, httpmock.NewStringResponder(status, body))
}

func getHwMockAdapter(t *testing.T) *adapter {
	hwRegistry := &model.Registry{
		ID:          1,
		Name:        "Huawei",
		Description: "Adapter for SWR -- The image registry of Huawei Cloud",
		Type:        model.RegistryTypeHuawei,
		URL:         "https://swr.cn-north-1.myhuaweicloud.com",
		Credential:  &model.Credential{AccessKey: "cn-north-1@IJYZLFBKBFN8LOUITAH", AccessSecret: "f31e8e2b948265afdae32e83722a7705fd43e154585ff69e64108247750e5d"},
		Insecure:    false,
		Status:      "",
	}
	adp, err := newAdapter(hwRegistry)
	if err != nil {
		t.Fatalf("Failed to call newAdapter(), reason=[%v]", err)
	}
	a := adp.(*adapter)

	httpmock.ActivateNonDefault(a.client.GetClient())
	httpmock.ActivateNonDefault(a.oriClient)
	t.Cleanup(httpmock.DeactivateAndReset)

	return a
}

func mockGetJwtToken(repository string) {
	httpmock.RegisterResponderWithQuery(http.MethodGet, mockHuaweiURL+"/swr/auth/v2/registry/auth",
		url.Values{"scope": []string{fmt.Sprintf("repository:%s:push,pull", repository)}},
		httpmock.NewJsonResponderOrPanic(http.StatusOK, jwtToken{Token: "token"}))
}

func TestAdapter_FetchArtifacts(t *testing.T) {
	httpmock.RegisterResponderWithQuery(http.MethodGet, mockHuaweiURL+"/dockyard/v2/repositories",
		url.Values{"filter": []string{"center::self"}},
		func(req *http.Request) (*http.Response, error) {
			username, password, ok := req.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "cn-north-1@IJYZLFBKBFN8LOUITAH", username)
			assert.Equal(t, "f31e8e2b948265afdae32e83722a7705fd43e154585ff69e64108247750e5d", password)
			return httpmock.NewJsonResponse(http.StatusOK, []hwRepoQueryResult{
				{Name: "name1"},
				{Name: "name2"},
			})
		})

	a := getHwMockAdapter(t)
	resources, err := a.FetchArtifacts(nil)
	assert.NoError(t, err)
	assert.Len(t, resources, 2)
}

func TestAdapter_ManifestExist(t *testing.T) {
	mockGetJwtToken("sundaymango_mango/hello-world")
	httpmock.RegisterResponder(http.MethodGet, mockHuaweiURL+"/v2/sundaymango_mango/hello-world/manifests/latest",
		httpmock.NewJsonResponderOrPanic(http.StatusOK, hwManifest{
			MediaType: distribution.ManifestMediaTypes()[0],
		}))

	a := getHwMockAdapter(t)
	exist, _, err := a.ManifestExist("sundaymango_mango/hello-world", "latest")
	assert.NoError(t, err)
	assert.True(t, exist)
}

func TestAdapter_DeleteManifest(t *testing.T) {
	mockGetJwtToken("sundaymango_mango/hello-world")
	registerHuaweiResponder(http.MethodDelete, "/v2/sundaymango_mango/hello-world/manifests/latest", http.StatusOK, "")

	a := getHwMockAdapter(t)
	err := a.DeleteManifest("sundaymango_mango/hello-world", "latest")
	assert.NoError(t, err)
}
