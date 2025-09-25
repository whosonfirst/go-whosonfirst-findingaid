package resolve

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
)

type RunOptions struct {
	ResolverURI string
	Ids         []int64
	Verbose     bool
}

func RunOptionsFromFlagSet(fs *flag.FlagSet) (*RunOptions, error) {

	flagset.Parse(fs)

	opts := &RunOptions{
		ResolverURI: resolver_uri,
		Ids:         ids,
		Verbose:     verbose,
	}

	return opts, nil
}
