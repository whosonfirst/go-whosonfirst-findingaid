package repo

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/whosonfirst/go-cache"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	"github.com/whosonfirst/go-whosonfirst-uri"
	_ "log"
	"net/url"
	"strconv"
)

// RepoResolver is a struct that implements the findingaid.Resolver interface for information about Who's On First repositories.
type RepoResolver struct {
	findingaid.Resolver
	cache cache.Cache
}

func init() {

	ctx := context.Background()
	err := findingaid.RegisterResolver(ctx, "repo", NewRepoResolver)

	if err != nil {
		panic(err)
	}
}

// NewRepoResolver returns a findingaid.Resolver instance for exposing information about Who's On First repositories
func NewRepoResolver(ctx context.Context, uri string) (findingaid.Resolver, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	cache_uri := q.Get("cache")

	if cache_uri == "" {
		return nil, errors.New("Missing cache URI")
	}

	_, err = url.Parse(cache_uri)

	if err != nil {
		return nil, err
	}

	c, err := cache.NewCache(ctx, cache_uri)

	if err != nil {
		return nil, err
	}

	fa := &RepoResolver{
		cache: c,
	}

	return fa, nil
}

// LookupID will return 'repo.ResolverResponse' for 'id' if it present in the finding aid.
func (fa *RepoResolver) Resolve(ctx context.Context, str_uri string) (interface{}, error) {

	id, _, err := uri.ParseURI(str_uri)

	if err != nil {
		return nil, err
	}

	str_id := strconv.FormatInt(id, 10)

	fh, err := fa.cache.Get(ctx, str_id)

	if err != nil {
		return nil, err
	}

	var rsp *FindingAidResponse

	dec := json.NewDecoder(fh)
	err = dec.Decode(&rsp)

	if err != nil {
		return nil, err
	}

	return rsp, nil
}