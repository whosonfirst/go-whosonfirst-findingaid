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
	"net/url"
	"strconv"
)

func init() {
	ctx := context.Background()
	wof_cache.RegisterCache(ctx, "readercache", NewReaderCache)
}

type ReaderCache struct {
	wof_cache.Cache
	reader reader.Reader
	cache  wof_cache.Cache
}

func NewReaderCache(ctx context.Context, uri string) (wof_cache.Cache, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	r_uri := q.Get("reader")

	if r_uri == "" {
		return nil, errors.New("Missing reader parameter")
	}

	r, err := reader.NewReader(ctx, r_uri)

	if err != nil {
		return nil, err
	}

	c_uri := q.Get("cache")

	if c_uri == "" {
		return nil, errors.New("Missing cache parameter")
	}

	c, err := wof_cache.NewCache(ctx, c_uri)

	if err != nil {
		return nil, err
	}

	rc := &ReaderCache{
		reader: r,
		cache:  c,
	}

	return rc, nil
}

func (c *ReaderCache) Name() string {
	return "readercache"
}

func (c *ReaderCache) Get(ctx context.Context, key string) (io.ReadCloser, error) {

	c_fh, err := c.cache.Get(ctx, key)

	if err != nil {
		return c_fh, nil
	}

	id, err := strconv.ParseInt(key, 10, 64)

	if err != nil {
		return nil, err
	}

	rel_path, err := uri.Id2RelPath(id)

	if err != nil {
		return nil, err
	}

	r_fh, err := c.reader.Read(ctx, rel_path)

	if err != nil {
		return nil, err
	}

	defer r_fh.Close()

	body, err := ioutil.ReadAll(r_fh)

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
	rsp := ioutil.NopCloser(br)

	return c.Set(ctx, key, rsp)
}

func (c *ReaderCache) Set(ctx context.Context, key string, fh io.ReadCloser) (io.ReadCloser, error) {
	return c.cache.Set(ctx, key, fh)
}

func (c *ReaderCache) Unset(ctx context.Context, key string) error {
	return c.cache.Unset(ctx, key)
}

func (c *ReaderCache) Hits() int64 {
	return c.cache.Hits()
}

func (c *ReaderCache) Misses() int64 {
	return c.cache.Misses()
}

func (c *ReaderCache) Evictions() int64 {
	return c.cache.Evictions()
}

func (c *ReaderCache) Size() int64 {
	return c.cache.Size()
}
