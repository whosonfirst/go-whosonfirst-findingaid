package repo

import (
	"context"
	"fmt"
	_ "github.com/whosonfirst/go-whosonfirst-index/fs"	
	"testing"
)

func TestFindingAid(t *testing.T) {

	ctx := context.Background()
	
	wof_id := int64(1444838459)

	repo := "whosonfirst-data-admin-is"	
	repo_url := fmt.Sprintf("../fixtures/%s", repo)
	
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

	if rsp.Repo != repo {
		t.Fatal("Invalid repo")
	}
}
