package git

import (
	"fmt"
	"context"
	"testing"
)

func TestGitFindingAid(t *testing.T) {

	repo := "whosonfirst-data-admin-is"
	repo_url := fmt.Sprintf("https://github.com/whosonfirst-data/%s.git", repo)

	wof_id := int64(1444838459)
	
	ctx := context.Background()

	fa, err := NewRepoFindingAid(ctx)

	if err != nil {
		t.Fatal(err)
	}

	err = fa.Index(ctx, repo_url)

	if err != nil {
		t.Fatal(err)
	}

	has_repo, err := fa.LookupID(ctx, wof_id)

	if err != nil {
		t.Fatal(err)
	}

	if has_repo != repo {
		t.Fatal("Invalid repo")
	}
}
