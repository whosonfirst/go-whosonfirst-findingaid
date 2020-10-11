package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aaronland/go-http-server"
	"github.com/rs/cors"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-cache"
	"github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-reader-http"
	"github.com/whosonfirst/go-whosonfirst-findingaid/http"
	"github.com/whosonfirst/go-whosonfirst-findingaid/repo"
	"github.com/whosonfirst/go-whosonfirst-index"
	"io"
	"log"
	go_http "net/http"
)

func init() {
	ctx := context.Background()
	cache.RegisterCache(ctx, "http", NewHTTPCache)

	ni := NewNullIndexer()
	index.Register("null", ni)
}

type NullIndexer struct {
	index.Driver
}

func NewNullIndexer() index.Driver {
	return &NullIndexer{}
}

func (i *NullIndexer) Open(string) error {
	return nil
}

func IndexURI(context.Context, index.IndexerFunc, string) error {
	return nil
}

type HTTPCache struct {
	cache.Cache
	reader reader.Reader
}

func NewHTTPCache(ctx context.Context, uri string) (cache.Cache, error) {

	r, err := reader.NewReader(ctx, "https://data.whosonfirst.org")

	if err != nil {
		return nil, err
	}

	c := &HTTPCache{
		reader: r,
	}

	return c, nil
}

func (c *HTTPCache) Name() string {
	return "http"
}

func (c *HTTPCache) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, errors.New("Not implemented")
}

func (c *HTTPCache) Set(ctx context.Context, key string, fh io.ReadCloser) (io.ReadCloser, error) {
	return fh, nil
}

func (c *HTTPCache) Unset(context.Context, string) error {
	return nil
}

func (c *HTTPCache) Hits() int64 {
	return 0
}

func (c *HTTPCache) Misses() int64 {
	return 0
}

func (c *HTTPCache) Evictions() int64 {
	return 0
}

func (c *HTTPCache) Size() int64 {
	return 0
}

func main() {

	fs := flagset.NewFlagSet("findingaid")

	server_uri := fs.String("server-uri", "http://localhost:8080", "A valid aaronland/go-http-server URI")

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVarsWithFeedback(fs, "FINDINGAID", true)

	if err != nil {
		log.Fatalf("Failed to set flags, %v", err)
	}

	ctx := context.Background()

	cache_uri := "null://"
	indexer_uri := "http://"

	fa_uri := fmt.Sprintf("repo:///?cache=%s&indexer=%s", cache_uri, indexer_uri)

	fa, err := repo.NewRepoFindingAid(ctx, fa_uri)

	if err != nil {
		log.Fatalf("Failed to create repo finding aid, %v", err)
	}

	cors_handler := cors.New(cors.Options{})

	lookup_handler, err := http.LookupHandler(fa)

	if err != nil {
		log.Fatalf("Failed to create lookup handler, %v", err)
	}

	lookup_handler = cors_handler.Handler(lookup_handler)

	mux := go_http.NewServeMux()

	mux.Handle("/", lookup_handler)

	s, err := server.NewServer(ctx, *server_uri)

	if err != nil {
		log.Fatalf("Failed to create server, %v", err)
	}

	log.Printf("Listening on %s", s.Address())

	err = s.ListenAndServe(ctx, mux)

	if err != nil {
		log.Fatalf("Failed to start server, %v", err)
	}
}
