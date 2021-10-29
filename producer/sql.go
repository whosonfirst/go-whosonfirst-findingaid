package producer

import (
	"context"
	gosql "database/sql"
	"fmt"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/producer/sql"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	_ "net/url"
	"sync"
)

type SQLProducer struct {
	Producer
	engine string
	db     *gosql.DB
}

func init() {
	ctx := context.Background()
	RegisterProducer(ctx, "sql", NewSQLProducer)
}

func NewSQLProducer(ctx context.Context, uri string) (Producer, error) {

	db, engine, err := sql.CreateDB(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create database, %w", err)
	}

	p := &SQLProducer{
		engine: engine,
		db:     db,
	}

	return p, nil
}

func (p *SQLProducer) PopulateWithIterator(ctx context.Context, monitor timings.Monitor, iterator_uri string, iterator_sources ...string) error {

	mu := new(sync.RWMutex)

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

		repo, exists, err := findingaid.GetRepoWithBytes(ctx, body)

		if err != nil {
			return fmt.Errorf("Failed to retrieve repo for %s, %w", path, err)
		}

		repo_id := repo.Id
		repo_name := repo.Name

		mu.Lock()
		defer mu.Unlock()

		if !exists {

			err = sql.AddToSources(ctx, p.db, repo_name, repo_id)

			if err != nil {
				return fmt.Errorf("Failed to store %s, %w", repo_name, err)
			}
		}

		err = sql.AddToCatalog(ctx, p.db, id, repo_id)
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

func (p *SQLProducer) Close(ctx context.Context) error {
	return nil
}
