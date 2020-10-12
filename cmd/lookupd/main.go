package main

import (
	"context"
	"github.com/aaronland/go-http-server"
	"github.com/rs/cors"
	"github.com/sfomuseum/go-flags/flagset"
	_ "github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-reader-http"
	"github.com/whosonfirst/go-whosonfirst-findingaid/http"
	_ "github.com/whosonfirst/go-whosonfirst-findingaid/index"
	"github.com/whosonfirst/go-whosonfirst-findingaid/repo"
	"log"
	go_http "net/http"
	"net/url"
)

func main() {

	fs := flagset.NewFlagSet("findingaid")

	server_uri := fs.String("server-uri", "http://localhost:8080", "A valid aaronland/go-http-server URI string.")

	cache_uri := fs.String("cache-uri", "readercache://?reader=http://data.whosonfirst.org&cache=gocache://", "A valid whosonfirst/go-cache URI string.")
	indexer_uri := fs.String("indexer-uri", "null://", "A valid whosonfirst/go-whosonfirst-index URI string.")

	findingaid_uri := fs.String("findingaid-uri", "repo://?cache={cache_uri}&indexer={indexer_uri}", "A valid whosonfirst/go-whosonfirst-findingaid URI string.")

	enable_cors := fs.Bool("enable-cors", true, "Enable CORS headers for output.")

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVarsWithFeedback(fs, "FINDINGAID", true)

	if err != nil {
		log.Fatalf("Failed to set flags, %v", err)
	}

	ctx := context.Background()

	fa_uri, err := url.Parse(*findingaid_uri)

	if err != nil {
		log.Fatalf("Failed to parse findingaid URI, %v", err)
	}

	fa_q := fa_uri.Query()

	if fa_q.Get("cache") == "{cache_uri}" {
		fa_q["cache"] = []string{*cache_uri}
	}

	if fa_q.Get("indexer") == "{indexer_uri}" {
		fa_q["indexer"] = []string{*indexer_uri}
	}

	fa_uri.RawQuery = fa_q.Encode()

	fa, err := repo.NewRepoFindingAid(ctx, fa_uri.String())

	if err != nil {
		log.Fatalf("Failed to create repo finding aid, %v", err)
	}

	lookup_handler, err := http.LookupHandler(fa)

	if err != nil {
		log.Fatalf("Failed to create lookup handler, %v", err)
	}

	if *enable_cors {
		cors_handler := cors.New(cors.Options{})
		lookup_handler = cors_handler.Handler(lookup_handler)
	}

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
