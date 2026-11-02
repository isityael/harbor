package handlers

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/docker/distribution/configuration"
)

func TestHarborVEXManifestPayloadBound(t *testing.T) {
	for _, chunked := range []bool{false, true} {
		for _, size := range []int{int(maxManifestBodySize), int(maxManifestBodySize) + 1} {
			r := httptest.NewRequest("PUT", "/v2/repo/manifests/tag", strings.NewReader(strings.Repeat("x", size)))
			if chunked {
				r.ContentLength = -1
			}
			var destination bytes.Buffer
			err := copyFullPayload(context.Background(), httptest.NewRecorder(), r, &destination, maxManifestBodySize, "manifest PUT regression")
			if (err != nil) != (size > int(maxManifestBodySize)) || destination.Len() > int(maxManifestBodySize) {
				t.Fatalf("size=%d chunked=%v stored=%d error=%v", size, chunked, destination.Len(), err)
			}
		}
	}
}

func TestHarborVEXCatalogAllocationBound(t *testing.T) {
	for _, n := range []string{"6", "9223372036854775807"} {
		ctx := &Context{App: &App{Config: &configuration.Configuration{Catalog: configuration.Catalog{MaxEntries: 5}}}, Context: context.Background()}
		handler := &catalogHandler{Context: ctx}
		handler.GetCatalog(httptest.NewRecorder(), httptest.NewRequest("GET", "/v2/_catalog?n="+n, nil))
		if len(ctx.Errors) != 1 {
			t.Fatalf("n=%s must reject before allocating or accessing registry, got %v", n, ctx.Errors)
		}
	}
	// The zero-page control must succeed even without a storage backend.
	ctx := &Context{App: &App{Config: &configuration.Configuration{Catalog: configuration.Catalog{MaxEntries: 5}}}, Context: context.Background()}
	response := httptest.NewRecorder()
	(&catalogHandler{Context: ctx}).GetCatalog(response, httptest.NewRequest("GET", "/v2/_catalog?n=0", nil))
	if len(ctx.Errors) != 0 || response.Code != 200 || !strings.Contains(response.Body.String(), `"repositories":[]`) {
		t.Fatalf("zero-page control failed: errors=%v body=%s", ctx.Errors, response.Body.String())
	}
}
