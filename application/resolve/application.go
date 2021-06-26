package resolve

import (
	"context"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	"github.com/whosonfirst/go-whosonfirst-findingaid/application"
	"net/url"
)

var cache_uri string
var findingaid_uri string

type ResolveApplication struct {
	application.Application
}

func NewResolveApplication(ctx context.Context) (application.Application, error) {
	app := &ResolveApplication{}
	return app, nil
}

func (app *ResolveApplication) DefaultFlagSet(ctx context.Context) (*flag.FlagSet, error) {

	fs := flagset.NewFlagSet("resolve")

	fs.StringVar(&cache_uri, "cache-uri", "file:///tmp", "A valid whosonfirst/go-cache URI string.")
	fs.StringVar(&findingaid_uri, "findingaid-uri", "repo://?cache={cache_uri}", "A valid whosonfirst/go-whosonfirst-findingaid URI string.")

	return fs, nil
}

func (app *ResolveApplication) Run(ctx context.Context) error {

	fs, err := app.DefaultFlagSet(ctx)

	if err != nil {
		return err
	}

	return app.RunWithFlagSet(ctx, fs)
}

func (app *ResolveApplication) RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

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

	uris := fs.Args()

	for _, uri := range uris {

		rsp, err := fa.ResolveURI(ctx, uri)

		if err != nil {
			return fmt.Errorf("Failed to resolve '%s', %v", uri, err)
		}

		fmt.Println(rsp)
	}

	return nil
}
