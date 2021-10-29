package producer

import (
	"context"
	"github.com/aaronland/go-roster"
	"github.com/sfomuseum/go-timings"
	"net/url"
)

type Producer interface {
	PopulateWithIterator(context.Context, timings.Monitor, string, ...string) error
	Close(context.Context) error
}

var producer_roster roster.Roster

type ProducerInitializationFunc func(ctx context.Context, uri string) (Producer, error)

func RegisterProducer(ctx context.Context, scheme string, init_func ProducerInitializationFunc) error {

	err := ensureProducerRoster()

	if err != nil {
		return err
	}

	return producer_roster.Register(ctx, scheme, init_func)
}

func ensureProducerRoster() error {

	if producer_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		producer_roster = r
	}

	return nil
}

func NewProducer(ctx context.Context, uri string) (Producer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := producer_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(ProducerInitializationFunc)
	return init_func(ctx, uri)
}
