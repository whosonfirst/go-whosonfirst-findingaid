package resolver

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

func TestReaderResolver(t *testing.T) {

	ctx := context.Background()

	fixtures := "../fixtures"
	abs_path, err := filepath.Abs(fixtures)

	if err != nil {
		t.Fatalf("Failed to derive absolute path for fixtures, %v", err)
	}

	reader_uri := fmt.Sprintf("fs://%s", abs_path)
	resolver_uri := fmt.Sprintf("reader://?reader=%s", reader_uri)

	r, err := NewResolver(ctx, resolver_uri)

	if err != nil {
		t.Fatalf("Failed to create resolver for '%s', %v", resolver_uri, err)
	}

	repo, err := r.GetRepo(ctx, 101736545)

	if err != nil {
		t.Fatalf("Failed to get repo, %v", err)
	}

	if repo != "whosonfirst-data-admin-ca" {
		t.Fatalf("Invalid repo: %s", repo)
	}

	fmt.Printf(repo)
}
