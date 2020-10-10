package main

import (
	"context"
	"github.com/aaronland/go-http-server"
	"github.com/rs/cors"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-cache"		
	_ "github.com/whosonfirst/go-reader-http"
	"github.com/whosonfirst/go-whosonfirst-findingaid/http"
	"log"
	"io"
	"errors"
	go_http "net/http"
)

type NullIndexer struct {
	index.Driver
}

func (i *NullIndexer) Open(string) error {
	return nil
}

func IndexURI(context.Context, index.IndexerFunc, string) error {
	return nil
}

type HTTPCache struct {
	cache.Cache
}

func (c *HTTPCache) Name() string {
	return "http"
}

func (c *HTTPCache) Get(key string) (io.ReadCloser, error) {
	return nil, errors.New("Not implemented")
}

func (c *HTTPCache) Set(key string, fh io.ReadCloser) (io.ReadCloser, error) {
	return fh, nil
}

func (c *HTTPCache) Unset(string) error {
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
	reader_uri := fs.String("reader-uri", "", "A valid whosonfirst/go-reader URI")

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVarsWithFeedback(fs, "FINDINGAID", true)

	if err != nil {
		log.Fatalf("Failed to set flags, %v", err)
	}

	ctx := context.Background()

	r, err := reader.NewReader(ctx, *reader_uri)

	if err != nil {
		log.Fatalf("Failed to create reader, %v", err)
	}

	cors_handler := cors.New(cors.Options{})

	lookup_handler, err := http.LookupHandler(r)

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
