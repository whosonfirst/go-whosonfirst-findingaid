package resolve

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/resolver"
)

func Run(ctx context.Context) error {
	fs := DefaultFlagSet()
	return RunWithFlagSet(ctx, fs)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

	opts, err := RunOptionsFromFlagSet(fs)

	if err != nil {
		return fmt.Errorf("Failed to derive run options, %w", err)
	}

	return RunWithOptions(ctx, opts)
}

func RunWithOptions(ctx context.Context, opts *RunOptions) error {

	if opts.Verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	r, err := resolver.NewResolver(ctx, opts.ResolverURI)

	if err != nil {
		return fmt.Errorf("Failed to create new resolver, %w", err)
	}

	for _, id := range opts.Ids {

		slog.Debug("Get repo", "id", id)
		repo, err := r.GetRepo(ctx, id)

		if err != nil {
			return fmt.Errorf("Failed to derive repo for %d, %w", id, err)
		}

		fmt.Printf("%d\t%s\n", id, repo)
	}

	return nil
}
