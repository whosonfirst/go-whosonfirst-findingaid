package repo

import (
	"context"
	"fmt"
	_ "github.com/whosonfirst/go-whosonfirst-index/fs"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"testing"
)

func TestFindingAid(t *testing.T) {

	ctx := context.Background()

	wof_id := int64(1444838459)

	wof_path, err := uri.Id2RelPath(wof_id)

	if err != nil {
		t.Fatal(err)
	}

	wof_repo := "whosonfirst-data-admin-is"
	repo_url := fmt.Sprintf("../fixtures/%s", wof_repo)

	cache_uri := "gocache://"
	indexer_uri := "repo://"

	fa_uri := fmt.Sprintf("repo:///?cache=%s&indexer=%s", cache_uri, indexer_uri)

	fa, err := NewRepoFindingAid(ctx, fa_uri)

	if err != nil {
		t.Fatal(err)
	}

	err = fa.Index(ctx, repo_url)

	if err != nil {
		t.Fatal(err)
	}

	var rsp FindingAidResponse

	err = fa.LookupID(ctx, wof_id, &rsp)

	if err != nil {
		t.Fatal(err)
	}

	if rsp.Repo != wof_repo {
		t.Fatal("Invalid repo")
	}

	if rsp.Path != wof_path {
		t.Fatal("Invalid path")
	}
}
