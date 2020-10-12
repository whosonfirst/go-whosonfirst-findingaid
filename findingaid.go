package findingaid

import (
	"context"
	"github.com/aaronland/go-roster"
	"io"
	"net/url"
)

type FindingAid interface {
	Index(context.Context, ...string) error
	IndexReader(context.Context, io.Reader) error
	LookupID(context.Context, int64) (interface{}, error)
}

type FindingAidInitializationFunc func(ctx context.Context, uri string) (FindingAid, error)

var findingaids roster.Roster

func ensureRoster() error {

	if findingaids == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		findingaids = r
	}

	return nil
}

func RegisterFindingAid(ctx context.Context, name string, c FindingAidInitializationFunc) error {

	err := ensureRoster()

	if err != nil {
		return err
	}

	return findingaids.Register(ctx, name, c)
}

func NewFindingAid(ctx context.Context, uri string) (FindingAid, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := findingaids.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init := i.(FindingAidInitializationFunc)
	c, err := init(ctx, uri)

	if err != nil {
		return nil, err
	}

	return c, nil
}
