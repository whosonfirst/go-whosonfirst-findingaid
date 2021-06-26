package repo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-cache"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	"github.com/whosonfirst/go-whosonfirst-iterate/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	_ "log"
	"net/url"
	"strconv"
)

// RepoFindingAid is a struct that implements the findingaid.FindingAid interface for information about Who's On First repositories.
type RepoFindingAid struct {
	findingaid.FindingAid
	cache        cache.Cache
	iterator_uri string
}

// FindingAidResonse is a struct that contains Who's On First repository information for Who's On First records.
type FindingAidResponse struct {
	// The unique Who's On First ID.
	ID int64 `json:"id"`
	// The name of the Who's On First repository.
	Repo string `json:"repo"`
	// The relative path for a Who's On First ID.
	URI string `json:"uri"`
}

func init() {

	ctx := context.Background()
	err := findingaid.RegisterFindingAid(ctx, "repo", NewRepoFindingAid)

	if err != nil {
		panic(err)
	}
}

// NewRepoFindingAid returns a findingaid.FindingAid instance for exposing information about Who's On First repositories
func NewRepoFindingAid(ctx context.Context, uri string) (findingaid.FindingAid, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	cache_uri := q.Get("cache")
	iterator_uri := q.Get("indexer")

	if cache_uri == "" {
		return nil, errors.New("Missing cache URI")
	}

	_, err = url.Parse(cache_uri)

	if err != nil {
		return nil, err
	}

	if iterator_uri == "" {
		return nil, errors.New("Missing indexer URI")
	}

	_, err = url.Parse(iterator_uri)

	if err != nil {
		return nil, err
	}

	c, err := cache.NewCache(ctx, cache_uri)

	if err != nil {
		return nil, err
	}

	fa := &RepoFindingAid{
		cache:        c,
		iterator_uri: iterator_uri,
	}

	return fa, nil
}

// Index will index records defined by 'sources...' in the finding aid, using the whosonfirst/go-whosonfirst-iterate package.
func (fa *RepoFindingAid) Index(ctx context.Context, sources ...string) error {

	cb := func(ctx context.Context, fh io.ReadSeeker, args ...interface{}) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		return fa.IndexReader(ctx, fh)
	}

	iter, err := iterator.NewIterator(ctx, fa.iterator_uri, cb)

	if err != nil {
		return err
	}

	return iter.IterateURIs(ctx, sources...)
}

// IndexReader will index an individual Who's On First record in the finding aid.
func (fa *RepoFindingAid) IndexReader(ctx context.Context, fh io.Reader) error {

	body, err := io.ReadAll(fh)

	if err != nil {
		return err
	}

	id_rsp := gjson.GetBytes(body, "properties.wof:id")

	if !id_rsp.Exists() {
		return errors.New("Missing wof:id")
	}

	repo_rsp := gjson.GetBytes(body, "properties.wof:repo")

	if !repo_rsp.Exists() {
		return errors.New("Missing wof:repo")
	}

	wof_id := id_rsp.Int()
	wof_repo := repo_rsp.String()

	rel_path, err := uri.Id2RelPath(wof_id)

	if err != nil {
		return err
	}

	rsp := &FindingAidResponse{
		ID:   wof_id,
		Repo: wof_repo,
		URI:  rel_path,
	}

	enc, err := json.Marshal(rsp)

	if err != nil {
		return err
	}

	br := bytes.NewReader(enc)
	br_cl := io.NopCloser(br)

	str_id := strconv.FormatInt(wof_id, 10)

	_, err = fa.cache.Set(ctx, str_id, br_cl)

	if err != nil {
		return err
	}

	return nil
}

// LookupID will return 'repo.FindingAidResponse' for 'id' if it present in the finding aid.
func (fa *RepoFindingAid) LookupID(ctx context.Context, id int64) (interface{}, error) {

	str_id := strconv.FormatInt(id, 10)

	fh, err := fa.cache.Get(ctx, str_id)

	if err != nil {
		return nil, err
	}

	var rsp *FindingAidResponse

	dec := json.NewDecoder(fh)
	err = dec.Decode(&rsp)

	if err != nil {
		return nil, err
	}

	return rsp, nil
}
