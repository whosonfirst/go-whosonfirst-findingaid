package provider

import (
	"context"
	"fmt"
	"github.com/jtacoma/uritemplates"
	"github.com/whosonfirst/go-whosonfirst-github/organizations"
	"github.com/whosonfirst/iso8601duration"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

type GitHubProvider struct {
	Provider
	org  string
	opts *organizations.ListOptions
}

func init() {
	ctx := context.Background()
	RegisterProvider(ctx, "github", NewGitHubProvider)
}

func NewGitHubProvider(ctx context.Context, uri string) (Provider, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	// u.Scheme is assumed to be github://
	org := u.Host

	opts := organizations.NewDefaultListOptions()

	q := u.Query()

	opts.Prefix = q["prefix"]
	opts.Exclude = q["exclude"]
	opts.AccessToken = q.Get("access_token")

	str_forked := q.Get("forked")
	str_not_forked := q.Get("not_forked")

	if str_forked != "" {

		forked, err := strconv.ParseBool(str_forked)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse 'forked' parameter, %w", err)
		}
		opts.Forked = forked
	}

	if str_not_forked != "" {

		not_forked, err := strconv.ParseBool(str_not_forked)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse 'not_forked' parameter, %w", err)
		}

		opts.NotForked = not_forked
	}

	updated_since := q.Get("updated_since")

	if updated_since != "" {

		var since time.Time

		is_timestamp, err := regexp.MatchString("^\\d+$", updated_since)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse updated_since time, %w", err)
		}

		if is_timestamp {

			ts, err := strconv.Atoi(updated_since)

			if err != nil {
				return nil, fmt.Errorf("Failed to parse updated_since parameter, %w", err)
			}

			now := time.Now()

			tm := time.Unix(int64(ts), 0)
			since = now.Add(-time.Since(tm))

		} else {

			// maybe also this https://github.com/araddon/dateparse ?

			d, err := duration.FromString(updated_since)

			if err != nil {
				return nil, fmt.Errorf("Failed to parse duration, %w", err)
			}

			now := time.Now()
			since = now.Add(-d.ToDuration())
		}

		// log.Printf("SINCE %v\n", since)
		// os.Exit(0)

		opts.PushedSince = &since
	}

	p := &GitHubProvider{
		org:  org,
		opts: opts,
	}

	return p, nil
}

func (p *GitHubProvider) IteratorSources(ctx context.Context) ([]string, error) {

	t := fmt.Sprintf("https://github.com/%s/{repo}.git", p.org)

	return p.IteratorSourcesWithURITemplate(ctx, t)
}

func (p *GitHubProvider) IteratorSourcesWithURITemplate(ctx context.Context, str_template string) ([]string, error) {

	t, err := uritemplates.Parse(str_template)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI template, %w", err)
	}

	repos, err := organizations.ListRepos(p.org, p.opts)

	if err != nil {
		return nil, fmt.Errorf("Failed to list repos, %w", err)
	}

	sources := make([]string, len(repos))

	for idx, name := range repos {

		values := map[string]interface{}{
			"repo": name,
		}

		uri, err := t.Expand(values)

		if err != nil {
			return nil, fmt.Errorf("Failed to expand template, %w", err)
		}

		sources[idx] = uri
	}

	return sources, nil
}
