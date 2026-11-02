package dockerhub

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/reg/model"
)

const (
	testUser     = ""
	testPassword = ""
)

func registerDockerHubResponder(method, path string, status int, body string) {
	httpmock.RegisterResponder(method, "https://hub.docker.com"+path,
		httpmock.NewStringResponder(status, body))
}

func getMockAdapter(t *testing.T) *adapter {
	r := &model.Registry{
		Type: model.RegistryTypeDockerHub,
		URL:  baseURL,
		Credential: &model.Credential{
			AccessKey:    testUser,
			AccessSecret: testPassword,
		},
	}
	ad, err := newAdapter(r)
	if err != nil {
		t.Fatalf("Failed to call newAdapter(), reason=[%v]", err)
	}
	a := ad.(*adapter)
	httpmock.ActivateNonDefault(a.client.client)
	t.Cleanup(httpmock.DeactivateAndReset)
	return a
}

func TestInfo(t *testing.T) {
	adapter := &adapter{}
	info, err := adapter.Info()
	require.Nil(t, err)
	require.Equal(t, 1, len(info.SupportedResourceTypes))
	assert.Equal(t, model.ResourceTypeImage, info.SupportedResourceTypes[0])
	assert.Equal(t, model.RepositoryPathComponentTypeOnlyTwo, info.SupportedRepositoryPathComponentType)
}

func TestListCandidateNamespaces(t *testing.T) {
	adapter := &adapter{}
	namespaces, err := adapter.listCandidateNamespaces("library/*")
	require.Nil(t, err)
	require.Equal(t, 1, len(namespaces))
	assert.Equal(t, "library", namespaces[0])
}

func TestListNamespaces(t *testing.T) {
	registerDockerHubResponder(http.MethodGet, "/v2/repositories/namespaces", http.StatusOK, "{}")

	a := getMockAdapter(t)

	namespaces, err := a.listNamespaces()
	assert.Nil(t, err)
	for _, ns := range namespaces {
		fmt.Println(ns)
	}
}

func TestFetchArtifacts(t *testing.T) {
	registerDockerHubResponder(http.MethodGet, "/v2/repositories/goharbor/", http.StatusOK, "{}")

	a := getMockAdapter(t)
	_, err := a.FetchArtifacts([]*model.Filter{
		{
			Type:  model.FilterTypeName,
			Value: "goharbor/harbor-core",
		},
	})
	require.Nil(t, err)
}
