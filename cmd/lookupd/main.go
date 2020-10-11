package main

import (
	"context"
	_ "fmt"
	"github.com/aaronland/go-http-server"
	"github.com/rs/cors"
	"github.com/sfomuseum/go-flags/flagset"
	_ "github.com/whosonfirst/go-whosonfirst-findingaid/cache"
	"github.com/whosonfirst/go-whosonfirst-findingaid/http"
	_ "github.com/whosonfirst/go-whosonfirst-findingaid/index"
	"github.com/whosonfirst/go-whosonfirst-findingaid/repo"
	"log"
	go_http "net/http"
	"net/url"
)

func main() {

	fs := flagset.NewFlagSet("findingaid")

	server_uri := fs.String("server-uri", "http://localhost:8080", "A valid aaronland/go-http-server URI")

	cache_uri := fs.String("cache-uri", "readercache://?reader=http://data.whosonfirst.org&cache=gocache://", "...")
	indexer_uri := fs.String("indexer-uri", "null://", "...")

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVarsWithFeedback(fs, "FINDINGAID", true)

	if err != nil {
		log.Fatalf("Failed to set flags, %v", err)
	}

	ctx := context.Background()

	fa_q := url.Values{}

	fa_q.Set("cache", *cache_uri)
	fa_q.Set("indexer", *indexer_uri)

	fa_uri := url.URL{}
	fa_uri.Scheme = "repo"
	fa_uri.RawQuery = fa_q.Encode()

	fa, err := repo.NewRepoFindingAid(ctx, fa_uri.String())

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
