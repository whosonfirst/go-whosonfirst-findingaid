package producer

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/sfomuseum/go-csvdict"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type CSVProducer struct {
	Producer
	archive_filename string
	catalog_filename string
	sources_filename string
	catalog_writer   io.WriteCloser
	sources_writer   io.WriteCloser
}

func init() {
	ctx := context.Background()
	RegisterProducer(ctx, "csv", NewCSVProducer)
}

func NewCSVProducer(ctx context.Context, uri string) (Producer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	archive_filename := q.Get("archive")

	if archive_filename == "" {
		return nil, fmt.Errorf("Missing ?archive parameter")
	}

	catalog_writer, err := ioutil.TempFile("", "catalog")

	if err != nil {
		return nil, fmt.Errorf("Failed to create catalog writer, %w", err)
	}

	sources_writer, err := ioutil.TempFile("", "sources")

	if err != nil {
		return nil, fmt.Errorf("Failed to create catalog writer, %w", err)
	}

	p := &CSVProducer{
		archive_filename: archive_filename,
		catalog_filename: catalog_writer.Name(),
		sources_filename: sources_writer.Name(),
		catalog_writer:   catalog_writer,
		sources_writer:   sources_writer,
	}

	return p, nil
}

func (p *CSVProducer) PopulateWithIterator(ctx context.Context, monitor timings.Monitor, iterator_uri string, iterator_sources ...string) error {

	var catalog_csv_wr *csvdict.Writer
	var sources_csv_wr *csvdict.Writer

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

		mu.Lock()
		defer mu.Unlock()

		if !exists {

			row := map[string]string{
				"id":   strconv.FormatInt(repo.Id, 10),
				"name": repo.Name,
			}

			if sources_csv_wr == nil {

				fieldnames := make([]string, 0)

				for k, _ := range row {
					fieldnames = append(fieldnames, k)
				}

				sort.Strings(fieldnames)

				w, err := csvdict.NewWriter(p.sources_writer, fieldnames)

				if err != nil {
					return fmt.Errorf("Failed to create CSV writer, %w", err)
				}

				err = w.WriteHeader()

				if err != nil {
					return fmt.Errorf("Failed to write CSV header, %w", err)
				}

				sources_csv_wr = w
			}

			err = sources_csv_wr.WriteRow(row)

			if err != nil {
				return fmt.Errorf("Failed to write row for %d, %w", id, err)
			}

			sources_csv_wr.Flush()

		}

		row := map[string]string{
			"id":      strconv.FormatInt(id, 10),
			"repo_id": strconv.FormatInt(repo.Id, 10),
		}

		if catalog_csv_wr == nil {

			fieldnames := make([]string, 0)

			for k, _ := range row {
				fieldnames = append(fieldnames, k)
			}

			sort.Strings(fieldnames)

			w, err := csvdict.NewWriter(p.catalog_writer, fieldnames)

			if err != nil {
				return fmt.Errorf("Failed to create CSV writer, %w", err)
			}

			err = w.WriteHeader()

			if err != nil {
				return fmt.Errorf("Failed to write CSV header, %w", err)
			}

			catalog_csv_wr = w
		}

		err = catalog_csv_wr.WriteRow(row)

		if err != nil {
			return fmt.Errorf("Failed to write row for %d, %w", id, err)
		}

		catalog_csv_wr.Flush()

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

	err = p.catalog_writer.Close()

	if err != nil {
		return fmt.Errorf("Failed to close catalog writer, %w", err)
	}

	err = p.sources_writer.Close()

	if err != nil {
		return fmt.Errorf("Failed to close sources writer, %w", err)
	}

	err = p.createArchive(ctx)

	return nil
}

func (p *CSVProducer) Close(ctx context.Context) error {

	paths := []string{
		p.catalog_filename,
		p.sources_filename,
	}

	for _, p := range paths {

		err := os.Remove(p)

		if err != nil {
			return fmt.Errorf("Failed to remove %s, %w", p, err)
		}
	}

	return nil
}

// Adapted from https://gist.github.com/maximilien/328c9ac19ab0a158a8df

func (p *CSVProducer) createArchive(ctx context.Context) error {

	f, err := os.OpenFile(p.archive_filename, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		return fmt.Errorf("Could not create archive file, %w", err)
	}

	defer f.Close()

	gzip_writer := gzip.NewWriter(f)
	defer gzip_writer.Close()

	tar_writer := tar.NewWriter(gzip_writer)
	defer tar_writer.Close()

	to_archive := []string{
		p.catalog_filename,
		p.sources_filename,
	}

	for _, path := range to_archive {

		err := p.addFileToTarWriter(ctx, path, tar_writer)

		if err != nil {
			return fmt.Errorf("Could not add file '%s', to archive, %w", path, err)
		}
	}

	return nil
}

func (p *CSVProducer) addFileToTarWriter(ctx context.Context, path string, tar_writer *tar.Writer) error {

	f, err := os.Open(path)

	if err != nil {
		return fmt.Errorf("Could not open file '%s', %w", path, err)
	}

	defer f.Close()

	stat, err := f.Stat()

	if err != nil {
		return fmt.Errorf("Could not get stat for file '%s', %w", path, err)
	}

	fname := filepath.Base(path)

	if strings.HasPrefix(fname, "catalog") {
		fname = "catalog.csv"
	}

	if strings.HasPrefix(fname, "sources") {
		fname = "sources.csv"
	}

	header := &tar.Header{
		Name:    fname,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	err = tar_writer.WriteHeader(header)

	if err != nil {
		return fmt.Errorf("Could not write header for file '%s',%w", path, err)
	}

	_, err = io.Copy(tar_writer, f)

	if err != nil {
		return fmt.Errorf("Could not copy the file '%s' data to the tarball,%w", path, err)
	}

	return nil
}
