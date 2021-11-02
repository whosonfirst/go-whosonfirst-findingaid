package findingaid

import (
	"context"
	"fmt"
	"github.com/aaronland/go-artisanal-integers"
	"github.com/aaronland/go-brooklynintegers-api"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"io"
	"sync"
)

// type FindingAidRepo is an internal representation of a WOF-style repository guaranteed to have a unique ID.
type FindingAidRepo struct {
	Id   int64
	Name string
}

var repo_sources *sync.Map
var repo_mu *sync.RWMutex

var bi_client artisanalinteger.Client

func init() {

	repo_sources = new(sync.Map)
	repo_mu = new(sync.RWMutex)

	bi_client = api.NewAPIClient()
}

func GetRepoWithReader(ctx context.Context, r io.ReadSeeker) (*FindingAidRepo, bool, error) {

	body, err := io.ReadAll(r)

	if err != nil {
		return nil, false, fmt.Errorf("Failed to read document, %w", err)
	}

	return GetRepoWithBytes(ctx, body)
}

func GetRepoWithBytes(ctx context.Context, body []byte) (*FindingAidRepo, bool, error) {

	repo_name, err := properties.Repo(body)

	if err != nil {
		return nil, false, fmt.Errorf("Failed to derive repo, %w", err)
	}

	return GetRepo(ctx, repo_name)
}

func GetRepoWithBytesForPath(ctx context.Context, body []byte, path string) (*FindingAidRepo, bool, error) {

	rsp := gjson.GetBytes(body, path)

	if !rsp.Exists() {
		return nil, false, fmt.Errorf("Path (%s) does not exist", path)
	}

	repo_name := rsp.String()

	return GetRepo(ctx, repo_name)
}

func GetRepo(ctx context.Context, repo_name string) (*FindingAidRepo, bool, error) {

	repo_mu.Lock()
	defer repo_mu.Unlock()

	var repo *FindingAidRepo

	v, exists := repo_sources.Load(repo_name)

	if !exists {

		// Do we really need a 64-bit integer for this? No, no we don't.
		// But we do need something that will be reliably unique across
		// disparate runs.

		new_id, err := bi_client.NextInt()

		if err != nil {
			return nil, false, fmt.Errorf("Failed to create ID for repo, %w", err)
		}

		repo = &FindingAidRepo{
			Id:   new_id,
			Name: repo_name,
		}

		repo_sources.Store(repo_name, repo)
	} else {
		repo = v.(*FindingAidRepo)
	}

	return repo, exists, nil
}
