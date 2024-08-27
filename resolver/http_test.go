package resolver

import (
	"context"
	"testing"
)

func TestHTTPResolver(t *testing.T) {

	ctx := context.Background()

	r, err := NewResolver(ctx, "https://static.sfomuseum.org/findingaid/")

	if err != nil {
		t.Fatalf("Failed to create new resolver, %v", err)
	}

	id := int64(102528325)
	repo, err := r.GetRepo(ctx, id)

	if err != nil {
		t.Fatalf("Failed to get repo for %d, %v", id, err)
	}

	if repo != "sfomuseum-data-whosonfirst" {
		t.Fatalf("Unexpected repo, %s", repo)
	}
}
