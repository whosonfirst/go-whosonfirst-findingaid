package repo

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	"testing"
)

func TestHTTPResolver(t *testing.T) {

	ctx := context.Background()

	r, err := findingaid.NewResolver(ctx, "repo-http://")

	if err != nil {
		t.Fatalf("Failed to create new resolver, %v", err)
	}

	str_uri := "102527513"

	fa_rsp, err := r.ResolveURI(ctx, str_uri)

	if err != nil {
		t.Fatalf("Failed to resolve '%s', %v", str_uri, err)
	}

	rsp := fa_rsp.(*FindingAidResponse)

	if rsp.Repo != "whosonfirst-data-admin-us" {
		t.Fatalf("Unexpected response: %s", rsp.Repo)
	}
}
