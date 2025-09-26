package resolver

import (
	"context"
	"net/url"
	"testing"
)

func TestMultiResolver(t *testing.T) {

	ctx := context.Background()

	resolver_uris := []string{
		"https://static.sfomuseum.org/findingaid?template=https://raw.githubusercontent.com/sfomuseum-data/{repo}/main/data/",
		"https://data.whosonfirst.org/findingaid?template=https://raw.githubusercontent.com/whosonfirst-data/{repo}/master/data/",
	}

	r_query := url.Values{}

	for _, uri := range resolver_uris {
		r_query.Set("resolver", uri)
	}

	r_uri := url.URL{}
	r_uri.Scheme = "multi"
	r_uri.RawQuery = r_query.Encode()

	r, err := NewResolver(ctx, r_uri.String())

	if err != nil {
		t.Fatalf("Failed to create new resolver for %s, %v", r_uri.String(), err)
	}

	id := int64(85865975)
	repo, err := r.GetRepo(ctx, id)

	if err != nil {
		t.Fatalf("Failed to get repo for %d, %v", id, err)
	}

	if repo != "whosonfirst-data-admin-us" {
		t.Fatalf("Unexpected repo, %s", repo)
	}
}
