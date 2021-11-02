package producer

import (
	"context"
	"fmt"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/producer/protobuf"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"google.golang.org/protobuf/proto"
	"io"
	"net/url"
	"os"
	"sync"
)

type ProtobufProducer struct {
	Producer
	protobuf_writer io.WriteCloser
	path_repo       string
}

func init() {
	ctx := context.Background()
	RegisterProducer(ctx, "protobuf", NewProtobufProducer)
}

func NewProtobufProducer(ctx context.Context, uri string) (Producer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	protobuf_filename := u.Path

	if protobuf_filename == "" {
		return nil, fmt.Errorf("Missing ?protobuf parameter")
	}

	protobuf_wr, err := os.OpenFile(protobuf_filename, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		return nil, fmt.Errorf("Failed to open %s, %v", protobuf_filename, err)
	}

	q := u.Query()

	path_repo := q.Get("path-repo")

	p := &ProtobufProducer{
		protobuf_writer: protobuf_wr,
		path_repo:       path_repo,
	}

	return p, nil
}

func (p *ProtobufProducer) PopulateWithIterator(ctx context.Context, monitor timings.Monitor, iterator_uri string, iterator_sources ...string) error {

	catalog := &protobuf.Catalog{}
	sources := &protobuf.Sources{}

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

		var repo *findingaid.FindingAidRepo
		var exists bool

		if p.path_repo != "" {
			repo, exists, err = findingaid.GetRepoWithBytesForPath(ctx, body, p.path_repo)
		} else {
			repo, exists, err = findingaid.GetRepoWithBytes(ctx, body)
		}

		if err != nil {
			return fmt.Errorf("Failed to retrieve repo for %s, %w", path, err)
		}

		mu.Lock()
		defer mu.Unlock()

		if !exists {

			protobuf_repo := &protobuf.Repo{
				Id:   repo.Id,
				Name: repo.Name,
			}

			sources.Repos = append(sources.Repos, protobuf_repo)
		}

		// Store the record

		rec := &protobuf.Record{
			Id:   id,
			Repo: repo.Id,
		}

		catalog.Records = append(catalog.Records, rec)

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

	fa := &protobuf.FindingAid{
		Sources: sources,
		Catalog: catalog,
	}

	// TBD: Can we marshal directly to p.protobuf_writer ?

	out, err := proto.Marshal(fa)

	if err != nil {
		return fmt.Errorf("Failed to marshal pb, %w", err)
	}

	_, err = p.protobuf_writer.Write(out)

	if err != nil {
		return fmt.Errorf("Failed to write pb, %w", err)
	}

	err = p.protobuf_writer.Close()

	if err != nil {
		return fmt.Errorf("Failed to close pb, %w", err)
	}

	return nil
}

func (p *ProtobufProducer) Close(ctx context.Context) error {
	return nil
}
