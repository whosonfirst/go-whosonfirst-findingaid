package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/tidwall/gjson"
	wof_cache "github.com/whosonfirst/go-cache"
	"github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-reader-http"
	"github.com/whosonfirst/go-whosonfirst-findingaid/repo"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	"io/ioutil"
	"strconv"
)

func init() {
	ctx := context.Background()
	wof_cache.RegisterCache(ctx, "reader", NewReaderCache)
}

type ReaderCache struct {
	wof_cache.Cache
	reader reader.Reader
}

func NewReaderCache(ctx context.Context, uri string) (wof_cache.Cache, error) {

	r, err := reader.NewReader(ctx, "https://data.whosonfirst.org")

	if err != nil {
		return nil, err
	}

	c := &ReaderCache{
		reader: r,
	}

	return c, nil
}

func (c *ReaderCache) Name() string {
	return "http"
}

func (c *ReaderCache) Get(ctx context.Context, key string) (io.ReadCloser, error) {

	id, err := strconv.ParseInt(key, 10, 64)

	if err != nil {
		return nil, err
	}

	rel_path, err := uri.Id2RelPath(id)

	if err != nil {
		return nil, err
	}

	fh, err := c.reader.Read(ctx, rel_path)

	if err != nil {
		return nil, err
	}

	defer fh.Close()

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return nil, err
	}

	repo_rsp := gjson.GetBytes(body, "properties.wof:repo")

	if !repo_rsp.Exists() {
		return nil, errors.New("Invalid WOF record")
	}

	wof_repo := repo_rsp.String()

	fa_rsp := repo.FindingAidResponse{
		ID:   id,
		URI:  rel_path,
		Repo: wof_repo,
	}

	enc, err := json.Marshal(fa_rsp)

	if err != nil {
		return nil, err
	}

	br := bytes.NewReader(enc)
	return ioutil.NopCloser(br), nil
}

func (c *ReaderCache) Set(ctx context.Context, key string, fh io.ReadCloser) (io.ReadCloser, error) {
	return fh, nil
}

func (c *ReaderCache) Unset(context.Context, string) error {
	return nil
}

func (c *ReaderCache) Hits() int64 {
	return 0
}

func (c *ReaderCache) Misses() int64 {
	return 0
}

func (c *ReaderCache) Evictions() int64 {
	return 0
}

func (c *ReaderCache) Size() int64 {
	return 0
}
