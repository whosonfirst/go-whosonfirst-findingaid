package repo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/whosonfirst/go-cache"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	"io/ioutil"
	_ "log"
	"net/url"
	"reflect"
	"strconv"
)

type RepoFindingAid struct {
	findingaid.FindingAid
	cache       cache.Cache
	indexer_uri string
}

type FindingAidResponse struct {
	ID   int64  `json:"id"`
	Repo string `json:"repo"`
	Path string `json:"path"`
}

type geojson_properties struct {
	ID   int64  `json:"wof:id"`
	Repo string `json:"wof:repo"`
}

type geojson_feature struct {
	Properties geojson_properties `json:"properties"`
}

func NewRepoFindingAid(ctx context.Context, uri string) (findingaid.FindingAid, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	cache_uri := q.Get("cache")
	indexer_uri := q.Get("indexer")

	if cache_uri == "" {
		return nil, errors.New("Missing cache URI")
	}

	_, err = url.Parse(cache_uri)

	if err != nil {
		return nil, err
	}

	if indexer_uri == "" {
		return nil, errors.New("Missing indexer URI")
	}

	_, err = url.Parse(indexer_uri)

	if err != nil {
		return nil, err
	}

	c, err := cache.NewCache(ctx, cache_uri)

	if err != nil {
		return nil, err
	}

	fa := &RepoFindingAid{
		cache:       c,
		indexer_uri: indexer_uri,
	}

	return fa, nil
}

func (fa *RepoFindingAid) Index(ctx context.Context, sources ...string) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	remaining := len(sources)

	err_ch := make(chan error)
	done_ch := make(chan bool)

	for _, s := range sources {

		go func(ctx context.Context, source string) {

			err := fa.indexSource(ctx, source)

			if err != nil {
				err_ch <- err
			}

			done_ch <- true

		}(ctx, s)
	}

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			return err
		default:
			// pass
		}
	}

	return nil
}

func (fa *RepoFindingAid) IndexReader(ctx context.Context, fh io.Reader) error {

	var f *geojson_feature

	dec := json.NewDecoder(fh)
	err := dec.Decode(&f)

	if err != nil {
		return err
	}

	path, err := uri.Id2RelPath(f.Properties.ID)

	if err != nil {
		return err
	}

	rsp := &FindingAidResponse{
		ID:   f.Properties.ID,
		Repo: f.Properties.Repo,
		Path: path,
	}

	enc, err := json.Marshal(rsp)

	if err != nil {
		return err
	}

	br := bytes.NewReader(enc)
	br_cl := ioutil.NopCloser(br)

	str_id := strconv.FormatInt(f.Properties.ID, 10)

	_, err = fa.cache.Set(ctx, str_id, br_cl)

	if err != nil {
		return err
	}

	return nil
}

func (fa *RepoFindingAid) indexSource(ctx context.Context, source string) error {

	cb := func(ctx context.Context, fh io.Reader, args ...interface{}) error {

		select {
		case <-ctx.Done():
			return nil
		default:
			// pass
		}

		return fa.IndexReader(ctx, fh)
	}

	idx, err := index.NewIndexer(fa.indexer_uri, cb)

	if err != nil {
		return err
	}

	return idx.Index(ctx, source)
}

func (fa *RepoFindingAid) LookupID(ctx context.Context, id int64, i interface{}) error {

	str_id := strconv.FormatInt(id, 10)

	fh, err := fa.cache.Get(ctx, str_id)

	if err != nil {
		return err
	}

	var rsp *FindingAidResponse

	dec := json.NewDecoder(fh)
	err = dec.Decode(&rsp)

	if err != nil {
		return err
	}

	v := reflect.ValueOf(i).Elem()

	if f := v.FieldByName("ID"); f.IsValid() {
		f.SetInt(rsp.ID)
	}

	if f := v.FieldByName("Repo"); f.IsValid() {
		f.SetString(rsp.Repo)
	}

	if f := v.FieldByName("Path"); f.IsValid() {
		f.SetString(rsp.Path)
	}

	return nil
}
