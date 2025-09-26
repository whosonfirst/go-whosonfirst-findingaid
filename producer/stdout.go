package producer

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2"
	"github.com/whosonfirst/go-whosonfirst-iterate/v3"
	"github.com/whosonfirst/go-whosonfirst-uri"
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

	iter, err := iterate.NewIterator(ctx, iterator_uri)

	if err != nil {
		return fmt.Errorf("Failed to create iterator, %w", err)
	}

	for rec, err := range iter.Iterate(ctx, iterator_sources...) {

		if err != nil {
			return err
		}

		defer rec.Body.Close()

		id, uri_args, err := uri.ParseURI(rec.Path)

		if err != nil {
			rec.Body.Close()
			return fmt.Errorf("Failed to parse %s, %w", rec.Path, err)
		}

		if uri_args.IsAlternate {
			rec.Body.Close()
			return nil
		}

		// Get wof:repo

		body, err := io.ReadAll(rec.Body)

		if err != nil {
			rec.Body.Close()
			return fmt.Errorf("Failed to read %s, %w", rec.Path, err)
		}

		repo, _, err := findingaid.GetRepoWithBytes(ctx, body)

		if err != nil {
			rec.Body.Close()
			return fmt.Errorf("Failed to retrieve repo for %s, %w", rec.Path, err)
		}

		repo_name := repo.Name

		fmt.Fprintf(os.Stdout, "%d %s\n", id, repo_name)

		if err != nil {
			rec.Body.Close()
			return fmt.Errorf("Failed to store %s, %w", rec.Path, err)
		}

		go monitor.Signal(ctx)
		rec.Body.Close()
	}

	return nil
}

func (p *StdoutProducer) Close(ctx context.Context) error {
	return nil
}
