package storage_test

import (
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
)

type aclTransport func(*http.Request) (*http.Response, error)

func (f aclTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestHarborTypedACLLists(t *testing.T) {
	for _, object := range []bool{false, true} {
		name := "default"
		if object {
			name = "object"
		}
		for _, response := range []struct {
			name   string
			body   string
			status int
			want   []storage.ACLRule
		}{
			{"entities-and-roles", `{"items":[{"entity":"user-alice","role":"READER"},null,{"role":"OWNER"},{"entity":"allUsers"},{"entity":"group-team","role":"OWNER"}]}`, http.StatusOK, []storage.ACLRule{{Entity: "user-alice", Role: "READER"}, {Entity: "group-team", Role: "OWNER"}}},
			{"empty", `{"items":[]}`, http.StatusOK, []storage.ACLRule{}},
			{"permission-denied", `{"error":{"code":403,"message":"forbidden"}}`, http.StatusForbidden, nil},
		} {
			t.Run(name+"/"+response.name, func(t *testing.T) {
				client := &http.Client{Transport: aclTransport(func(request *http.Request) (*http.Response, error) {
					wantPath := "/storage/v1/b/example-bucket/defaultObjectAcl"
					if object {
						wantPath = "/storage/v1/b/example-bucket/o/example-object/acl"
					}
					if request.Method != http.MethodGet || request.URL.Path != wantPath {
						t.Fatalf("unexpected ACL request: %s %s", request.Method, request.URL.Path)
					}
					return &http.Response{StatusCode: response.status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(response.body)), Request: request}, nil
				})}
				ctx := cloud.NewContext("example-project", client)
				var got []storage.ACLRule
				var err error
				if object {
					got, err = storage.ACL(ctx, "example-bucket", "example-object")
				} else {
					got, err = storage.DefaultACL(ctx, "example-bucket")
				}
				if response.status != http.StatusOK {
					if err == nil || !strings.Contains(err.Error(), "forbidden") || got != nil {
						t.Fatalf("permission failure must be preserved: ACL=%v, error=%v", got, err)
					}
					return
				}
				if err != nil || !reflect.DeepEqual(got, response.want) {
					t.Fatalf("ACL=%v, error=%v; want %v", got, err, response.want)
				}
			})
		}
	}
}
