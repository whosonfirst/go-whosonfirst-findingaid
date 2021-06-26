package lookupd

import (
	"context"
	"flag"
	"fmt"
	"github.com/aaronland/go-http-server"
	"github.com/rs/cors"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	"github.com/whosonfirst/go-whosonfirst-findingaid/application"
	"github.com/whosonfirst/go-whosonfirst-findingaid/www"
	"log"
	"net/http"
	"net/url"
)

var server_uri string
var cache_uri string
var findingaid_uri string
var enable_cors bool

type LookupdApplication struct {
	application.Application
}

func NewLookupdApplication(ctx context.Context) (application.Application, error) {
	app := &LookupdApplication{}
	return app, nil
}

func (app *LookupdApplication) DefaultFlagSet(ctx context.Context) (*flag.FlagSet, error) {

	fs := flagset.NewFlagSet("lookupd")

	fs.StringVar(&server_uri, "server-uri", "http://localhost:8080", "A valid aaronland/go-http-server URI string.")

	fs.StringVar(&cache_uri, "cache-uri", "multi://?cache=gocache://&cache=file:///tmp", "A valid whosonfirst/go-cache URI string.")
	fs.StringVar(&findingaid_uri, "findingaid-uri", "repo://?cache={cache_uri}", "A valid whosonfirst/go-whosonfirst-findingaid URI string.")

	fs.BoolVar(&enable_cors, "enable-cors", true, "Enable CORS headers for output.")

	return fs, nil
}

func (app *LookupdApplication) Run(ctx context.Context) error {

	fs, err := app.DefaultFlagSet(ctx)

	if err != nil {
		return err
	}

	return app.RunWithFlagSet(ctx, fs)
}

func (app *LookupdApplication) RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVarsWithFeedback(fs, "FINDINGAID", true)

	if err != nil {
		return fmt.Errorf("Failed to set flags from environment variables, %w", err)
	}

	fa_uri, err := url.Parse(findingaid_uri)

	if err != nil {
		return fmt.Errorf("Failed to parse findingaid URI, %v", err)
	}

	fa_q := fa_uri.Query()

	if fa_q.Get("cache") == "{cache_uri}" {
		fa_q["cache"] = []string{cache_uri}
	}

	fa_uri.RawQuery = fa_q.Encode()

	fa, err := findingaid.NewResolver(ctx, fa_uri.String())

	if err != nil {
		return fmt.Errorf("Failed to create finding aid, %v", err)
	}

	lookup_handler, err := www.ResolveHandler(fa)

	if err != nil {
		return fmt.Errorf("Failed to create lookup handler, %v", err)
	}

	if enable_cors {
		cors_handler := cors.New(cors.Options{})
		lookup_handler = cors_handler.Handler(lookup_handler)
	}

	mux := http.NewServeMux()

	mux.Handle("/", lookup_handler)

	s, err := server.NewServer(ctx, server_uri)

	if err != nil {
		return fmt.Errorf("Failed to create server, %v", err)
	}

	log.Printf("Listening on %s", s.Address())

	err = s.ListenAndServe(ctx, mux)

	if err != nil {
		return fmt.Errorf("Failed to start server, %v", err)
	}

	return nil
}
