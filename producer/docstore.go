package producer

import (
	"context"
	"fmt"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/producer/docstore"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	gc_docstore "gocloud.dev/docstore"
	"io"
	"net/url"
)

/*

> java -Djava.library.path=./DynamoDBLocal_lib -jar DynamoDBLocal.jar
Initializing DynamoDB Local with the following configuration:
Port:	8000
InMemory:	false
DbPath:	null
SharedDb:	false
shouldDelayTransientStatuses:	false
CorsParams:	*

*/

type DocstoreProducer struct {
	Producer
	engine     string
	collection *gc_docstore.Collection
	path_repo  string
}

func init() {
	// ctx := context.Background()

	// load docstore things here
}

func NewDocstoreProducer(ctx context.Context, uri string) (Producer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	path_repo := q.Get("path-repo")
	q.Del("path-repo")

	u.RawQuery = q.Encode()

	uri = u.String()

	collection, err := gc_docstore.OpenCollection(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create collection, %w", err)
	}

	p := &DocstoreProducer{
		collection: collection,
		path_repo:  path_repo,
	}

	return p, nil
}

func (p *DocstoreProducer) PopulateWithIterator(ctx context.Context, monitor timings.Monitor, iterator_uri string, iterator_sources ...string) error {

	iter_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		id, uri_args, err := uri.ParseURI(path)

		if err != nil {
			return fmt.Errorf("Failed to parse %s, %w", path, err)
		}

		if uri_args.IsAlternate {
			return nil
		}

		// Get wof:repo

		body, err := io.ReadAll(fh)

		if err != nil {
			return fmt.Errorf("Failed to read %s, %w", path, err)
		}

		var repo *findingaid.FindingAidRepo

		if p.path_repo != "" {
			repo, _, err = findingaid.GetRepoWithBytesForPath(ctx, body, p.path_repo)
		} else {
			repo, _, err = findingaid.GetRepoWithBytes(ctx, body)
		}

		if err != nil {
			return fmt.Errorf("Failed to retrieve repo for %s, %w", path, err)
		}

		repo_name := repo.Name

		err = docstore.AddToCatalog(ctx, p.collection, id, repo_name)

		if err != nil {
			return fmt.Errorf("Failed to store %s, %w", path, err)
		}

		go monitor.Signal(ctx)
		return nil
	}

	iter, err := iterator.NewIterator(ctx, iterator_uri, iter_cb)

	if err != nil {
		return fmt.Errorf("Failed to create iterator, %v", err)
	}

	err = iter.IterateURIs(ctx, iterator_sources...)

	if err != nil {
		return fmt.Errorf("Failed to iterate sources, %v", err)
	}

	return nil
}

func (p *DocstoreProducer) Close(ctx context.Context) error {
	return p.collection.Close()
}
