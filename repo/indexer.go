package repo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-cache"
	"github.com/whosonfirst/go-ioutil"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	"github.com/whosonfirst/go-whosonfirst-iterate/iterator"
	"io"
	_ "log"
	"net/url"
)

// Indexer is a struct that implements the findingaid.Indexer interface for information about Who's On First repositories.
type Indexer struct {
	findingaid.Indexer
	cache         cache.Cache
	iterator_uri  string
	repo_property string
}

func init() {

	ctx := context.Background()
	err := findingaid.RegisterIndexer(ctx, "repo", NewIndexer)

	if err != nil {
		panic(err)
	}
}

// NewIndexer returns a findingaid.Indexer instance for exposing information about Who's On First repositories
func NewIndexer(ctx context.Context, uri string) (findingaid.Indexer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	cache_uri := q.Get("cache")
	iterator_uri := q.Get("iterator")

	if cache_uri == "" {
		return nil, errors.New("Missing ?cache= parameter.")
	}

	c, err := cache.NewCache(ctx, cache_uri)

	if err != nil {
		return nil, err
	}

	if iterator_uri == "" {
		return nil, errors.New("Missing ?iterator= parameter.")
	}

	// We defer creating the iterator until the 'IndexURIs' method is
	// invoked because the iterator callback has a reference to this
	// (findingaid indexer) instance which hasn't been created at this
	// point.

	_, err = url.Parse(iterator_uri)

	if err != nil {
		return nil, fmt.Errorf("Invalid ?iterator= parameter, %w", err)
	}

	// This is necessary for some repos like sfomuseum-data/sfomuseum-data-whosonfirst
	// where the relevant repo name is stored in a properties.sfomuseum:repo key. This
	// logic is handled below in IndexReader

	repo_property := WOF_REPO_PROPERTY

	custom_repo := q.Get("repo-property")

	if custom_repo != "" {
		repo_property = custom_repo
	}

	fa := &Indexer{
		cache:         c,
		iterator_uri:  iterator_uri,
		repo_property: repo_property,
	}

	return fa, nil
}

// Index will index records defined by 'sources...' in the finding aid, using the whosonfirst/go-whosonfirst-iterate package.
func (fa *Indexer) IndexURIs(ctx context.Context, sources ...string) error {

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
func (fa *Indexer) IndexReader(ctx context.Context, fh io.Reader) error {

	body, err := io.ReadAll(fh)

	if err != nil {
		return fmt.Errorf("Failed to read feature, %v", err)
	}

	// This is necessary for some repos like sfomuseum-data/sfomuseum-data-whosonfirst
	// where the relevant repo name is stored in a properties.sfomuseum:repo key

	if fa.repo_property != WOF_REPO_PROPERTY {

		path_wof_repo := fmt.Sprintf("properties.%s", WOF_REPO_PROPERTY)
		path_custom_repo := fmt.Sprintf("properties.%s", fa.repo_property)

		custom_rsp := gjson.GetBytes(body, path_custom_repo)

		// If custom repo property is present then use its value to update wof:repo

		if custom_rsp.Exists() {

			custom_repo := custom_rsp.String()

			body, err = sjson.SetBytes(body, path_wof_repo, custom_repo)

			if err != nil {
				return fmt.Errorf("Failed to assign custom repo (%s) to %s, %v", custom_repo, WOF_REPO_PROPERTY, err)
			}
		}
	}

	rsp, err := FindingAidResponseFromBytes(ctx, body)

	if err != nil {
		return err
	}

	enc, err := json.Marshal(rsp)

	if err != nil {
		return err
	}

	br := bytes.NewReader(enc)
	rsc, err := ioutil.NewReadSeekCloser(br)

	if err != nil {
		return err
	}

	key, err := cacheKeyFromRelPath(rsp.URI)

	if err != nil {
		return err
	}

	_, err = fa.cache.Set(ctx, key, rsc)

	if err != nil {
		return err
	}

	return nil
}
