// package provider interfaces used to generate a list of iterator `whosonfirst/go-whosonfirst-iterate/v2` URIs for crawling by a `producer.Producer` instance.
package provider

import (
	"context"
	"github.com/aaronland/go-roster"
	"net/url"
)

type Provider interface {
	IteratorSources(context.Context) ([]string, error)
	IteratorSourcesWithURITemplate(context.Context, string) ([]string, error)
}

var provider_roster roster.Roster

type ProviderInitializationFunc func(ctx context.Context, uri string) (Provider, error)

func RegisterProvider(ctx context.Context, scheme string, init_func ProviderInitializationFunc) error {

	err := ensureProviderRoster()

	if err != nil {
		return err
	}

	return provider_roster.Register(ctx, scheme, init_func)
}

func ensureProviderRoster() error {

	if provider_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		provider_roster = r
	}

	return nil
}

func NewProvider(ctx context.Context, uri string) (Provider, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := provider_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(ProviderInitializationFunc)
	return init_func(ctx, uri)
}
