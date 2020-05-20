package repo

import (
	"context"
	"fmt"
	_ "github.com/whosonfirst/go-whosonfirst-index-git"
	_ "log"
	"testing"
)

func TestGitFindingAid(t *testing.T) {

	repo := "whosonfirst-data-admin-is"
	repo_url := fmt.Sprintf("https://github.com/whosonfirst-data/%s.git", repo)

	wof_id := int64(1444838459)

	ctx := context.Background()

	fa, err := NewRepoFindingAid(ctx, "repo:///?cache=gocache://&indexer=git://")

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
