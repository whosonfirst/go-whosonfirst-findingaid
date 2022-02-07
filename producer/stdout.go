package producer

import (
	"context"
	"fmt"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	"os"
	"sync"
)

type StdoutProducer struct {
	Producer
}

func init() {
	ctx := context.Background()
	RegisterProducer(ctx, "stdout", NewStdoutProducer)
}

func NewStdoutProducer(ctx context.Context, uri string) (Producer, error) {

	p := &StdoutProducer{}
	return p, nil
}

func (p *StdoutProducer) PopulateWithIterator(ctx context.Context, monitor timings.Monitor, iterator_uri string, iterator_sources ...string) error {

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

		mu.Lock()
		defer mu.Unlock()

		repo, _, err := findingaid.GetRepoWithBytes(ctx, body)

		if err != nil {
			return fmt.Errorf("Failed to retrieve repo for %s, %w", path, err)
		}

		repo_name := repo.Name

		fmt.Fprintf(os.Stdout, "%d %s\n", id, repo_name)

		if err != nil {
			return fmt.Errorf("Failed to store %s, %w", path, err)
		}

		go monitor.Signal(ctx)
		return nil
	}

	iter, err := iterator.NewIterator(ctx, iterator_uri, iter_cb)

	if err != nil {
		return fmt.Errorf("Failed to create iterator, %w", err)
	}

	err = iter.IterateURIs(ctx, iterator_sources...)

	if err != nil {
		return fmt.Errorf("Failed to iterate sources, %w", err)
	}

	return nil
}

func (p *StdoutProducer) Close(ctx context.Context) error {
	return nil
}
