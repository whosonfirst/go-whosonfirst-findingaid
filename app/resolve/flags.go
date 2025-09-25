package resolve

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
)

var resolver_uri string
var ids multi.MultiInt64
var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("resolve")

	fs.StringVar(&resolver_uri, "resolver-uri", "", "...")
	fs.Var(&ids, "id", "One or more IDs to resolve")
	fs.BoolVar(&verbose, "verbose", false, "Enable verbose (debug) logging.")
	return fs
}
