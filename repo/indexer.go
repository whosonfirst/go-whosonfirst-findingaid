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

// RepoIndexer is a struct that implements the findingaid.Indexer interface for information about Who's On First repositories.
type RepoIndexer struct {
	findingaid.Indexer
	cache        cache.Cache
	iterator_uri string
}

func init() {

	ctx := context.Background()
	err := findingaid.RegisterIndexer(ctx, "repo", NewRepoIndexer)

	if err != nil {
		panic(err)
	}
}

// NewRepoIndexer returns a findingaid.Indexer instance for exposing information about Who's On First repositories
func NewRepoIndexer(ctx context.Context, uri string) (findingaid.Indexer, error) {

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

	_, err = url.Parse(iterator_uri)

	if err != nil {
		return nil, err
	}

	c, err := cache.NewCache(ctx, cache_uri)

	if err != nil {
		return nil, err
	}

	fa := &RepoIndexer{
		cache:        c,
		iterator_uri: iterator_uri,
	}

	return fa, nil
}

// Index will index records defined by 'sources...' in the finding aid, using the whosonfirst/go-whosonfirst-iterate package.
func (fa *RepoIndexer) IndexURI(ctx context.Context, sources ...string) error {

	if fa.iterator_uri == "" {
		return errors.New("Finding aid was not created with an indexer URI.")
	}

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
func (fa *RepoIndexer) IndexReader(ctx context.Context, fh io.Reader) error {

	body, err := io.ReadAll(fh)

	if err != nil {
		return err
	}

	// TO DO: SUPPORT ALT FILES

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
