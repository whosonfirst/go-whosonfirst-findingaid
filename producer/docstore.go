package producer

import (
	"context"
	"fmt"
	"github.com/aaronland/go-aws-dynamodb"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/producer/docstore"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	gc_docstore "gocloud.dev/docstore"
	gc_dynamodb "gocloud.dev/docstore/awsdynamodb"
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

	ctx := context.Background()

	RegisterProducer(ctx, "awsdynamodb", NewDocstoreProducer)

	for _, scheme := range gc_docstore.DefaultURLMux().CollectionSchemes() {

		err := RegisterProducer(ctx, scheme, NewDocstoreProducer)

		if err != nil {
			panic(err)
		}
	}
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

	// START OF put me in a package function or something

	var collection *gc_docstore.Collection

	if u.Scheme == "awsdynamodb" {

		// Connect local dynamodb using Golang
		// https://gist.github.com/Tamal/02776c3e2db7eec73c001225ff52e827
		// https://gocloud.dev/howto/docstore/#dynamodb-ctor

		client, err := dynamodb.NewClientWithURI(ctx, uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to create client, %v", err)
		}

		u, _ := url.Parse(uri)
		table_name := u.Host

		/*

			partition_key := q.Get("partition_key")

			// START OF necessary for order by created/lastupdate dates
			// https://pkg.go.dev/gocloud.dev@v0.23.0/docstore/awsdynamodb#InMemorySortFallback

			create_func := func() interface{} {
				return new(map[string]interface{})
			}

			fallback_func := aws_dynamodb.InMemorySortFallback(create_func)

			opts := &aws_dynamodb.Options{
				AllowScans:       true,
				RunQueryFallback: fallback_func,
			}

			// END OF necessary for order by created/lastupdate dates

			col, err := gc_dynamodb.OpenCollection(dynamodb.New(sess), table, partition_key, "", opts)

		*/

		col, err := gc_dynamodb.OpenCollection(client, table_name, "", "", nil)

		if err != nil {
			return nil, fmt.Errorf("Failed to open collection, %w", err)
		}

		collection = col

	} else {

		col, err := gc_docstore.OpenCollection(ctx, uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to create database for '%s', %w", uri, err)
		}

		collection = col
	}

	// END OF put me in a package function or something

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
