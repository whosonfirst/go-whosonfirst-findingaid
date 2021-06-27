package repo

import (
	"context"
	"encoding/json"
	"github.com/jtacoma/uritemplates"
	"github.com/whosonfirst/go-cache"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
	"github.com/whosonfirst/go-whosonfirst-uri"
	_ "log"
	"net/http"
	"net/url"
)

const FINDINGAID_URI_TEMPLATE string = "https://data.whosonfirst.org/findingaid/{id}"

type HTTPResolver struct {
	findingaid.Resolver
	cache    cache.Cache
	template *uritemplates.UriTemplate
}

func init() {

	ctx := context.Background()
	err := findingaid.RegisterResolver(ctx, "repo-http", NewHTTPResolver)

	if err != nil {
		panic(err)
	}
}

func NewHTTPResolver(ctx context.Context, uri string) (findingaid.Resolver, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	c, err := cache.NewCache(ctx, "gocache://")

	if err != nil {
		return nil, err
	}

	fa_template_uri := q.Get("findingaid_uri_template")

	if fa_template_uri == "" {
		fa_template_uri = FINDINGAID_URI_TEMPLATE
	}

	fa_template, err := uritemplates.Parse(fa_template_uri)

	if err != nil {
		return nil, err
	}

	fa := &HTTPResolver{
		cache:    c,
		template: fa_template,
	}

	return fa, nil
}

func (fa *HTTPResolver) ResolveURI(ctx context.Context, str_uri string) (interface{}, error) {

	id, _, err := uri.ParseURI(str_uri)

	if err != nil {
		return nil, err
	}

	values := map[string]interface{}{
		"id": id,
	}

	uri, err := fa.template.Expand(values)

	if err != nil {
		return nil, err
	}

	rsp, err := http.Get(uri)

	if err != nil {
		return "", err
	}

	defer rsp.Body.Close()

	var fa_rsp *FindingAidResponse

	dec := json.NewDecoder(rsp.Body)
	err = dec.Decode(&fa_rsp)

	if err != nil {
		return nil, err
	}

	return fa_rsp, nil
}
