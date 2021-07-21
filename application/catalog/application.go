package catalog

import (
	"context"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	"github.com/whosonfirst/go-whosonfirst-findingaid/application"
	_ "log"
	"net/url"
)

var cache_uri string
var iterator_uri string
var findingaid_uri string

type CatalogApplication struct {
	application.Application
}

func NewCatalogApplication(ctx context.Context) (application.Application, error) {
	app := &CatalogApplication{}
	return app, nil
}

func (app *CatalogApplication) DefaultFlagSet(ctx context.Context) (*flag.FlagSet, error) {

	fs := flagset.NewFlagSet("catalog")

	fs.StringVar(&cache_uri, "cache-uri", "file:///tmp/", "A valid whosonfirst/go-cache URI string.")
	fs.StringVar(&iterator_uri, "iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate URI string.")

	fs.StringVar(&findingaid_uri, "findingaid-uri", "repo://?cache={cache_uri}&iterator={iterator_uri}", "A valid whosonfirst/go-whosonfirst-findingaid URI string.")

	return fs, nil
}

func (app *CatalogApplication) Run(ctx context.Context) error {

	fs, err := app.DefaultFlagSet(ctx)

	if err != nil {
		return err
	}

	return app.RunWithFlagSet(ctx, fs)
}

func (app *CatalogApplication) RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

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

	if fa_q.Get("iterator") == "{iterator_uri}" {
		fa_q["iterator"] = []string{iterator_uri}
	}

	if fa_q.Get("iterator") == "" {
		return fmt.Errorf("Missing '-iterator-uri' flag.")
	}

	fa_uri.RawQuery = fa_q.Encode()

	fa, err := findingaid.NewIndexer(ctx, fa_uri.String())

	if err != nil {
		return fmt.Errorf("Failed to create finding aid, %v", err)
	}

	uris := fs.Args()

	err = fa.IndexURIs(ctx, uris...)

	if err != nil {
		return fmt.Errorf("Failed to catalog sources, %v", err)
	}

	return nil
}
