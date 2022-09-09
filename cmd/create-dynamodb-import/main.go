// create-dynamodb-import is a command-line tool to create a CSV file derived by one or more whosonfirst-findingaid data archives
// suitable for importing in to DynamoDB. For example:
//
//	$> go run cmd/create-dynamodb-import/main.go /usr/local/whosonfirst/whosonfirst-findingaids/data/*
//
// See also: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/S3DataImport.HowItWorks.html
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-csvdict"
	"io"
	"log"
	"os"
	"sync"
)

const sources_csv string = "sources.csv"
const catalog_csv string = "catalog.csv"

func main() {

	flag.Parse()

	writers := []io.Writer{
		os.Stdout,
	}

	wr := io.MultiWriter(writers...)

	fieldnames := []string{
		"id",
		"repo",
	}

	csv_wr, err := csvdict.NewWriter(wr, fieldnames)

	if err != nil {
		log.Fatalf("Failed to create CSV writer, %w", err)
	}

	err = csv_wr.WriteHeader()

	if err != nil {
		log.Fatalf("Failed to write CSV header, %w", err)
	}

	paths := flag.Args()

	done_ch := make(chan bool)
	err_ch := make(chan error)
	row_ch := make(chan map[string]string)

	for _, path := range paths {

		go func(path string) {

			defer func() {
				done_ch <- true
			}()

			r, err := os.Open(path)

			if err != nil {
				err_ch <- fmt.Errorf("Failed to open %s for reading, %w", path, err)
				return
			}

			defer r.Close()

			gr, err := gzip.NewReader(r)

			if err != nil {
				err_ch <- fmt.Errorf("Failed to create gzip reader for %s, %w", path, err)
				return
			}

			tr := tar.NewReader(gr)

			var catalog []byte
			var sources []byte

			for {
				hdr, err := tr.Next()

				if err == io.EOF {
					break // End of archive
				}

				if err != nil {
					err_ch <- fmt.Errorf("Failed to advance, %w", err)
					return
				}

				switch hdr.Name {
				case sources_csv, catalog_csv:
					// pass
				default:
					continue
				}

				body, err := io.ReadAll(tr)

				if err != nil {
					err_ch <- fmt.Errorf("Failed to read body for %s (%s), %w", path, hdr.Name, err)
					return
				}

				switch hdr.Name {
				case sources_csv:
					sources = body
				case catalog_csv:
					catalog = body
				default:
					// pass
				}
			}

			if len(sources) == 0 {
				return
			}

			sources_map := make(map[string]string)

			source_r, err := csvdict.NewReader(bytes.NewReader(sources))

			if err != nil {
				err_ch <- fmt.Errorf("Failed to create CSV reader for sources (%s), %w", path, err)
				return
			}

			for {
				row, err := source_r.Read()

				if err == io.EOF {
					break
				}

				if err != nil {
					err_ch <- fmt.Errorf("Failed to read CSV row for sources (%s), %w", path, err)
					return
				}

				sources_map[row["id"]] = row["name"]
			}

			catalog_r, err := csvdict.NewReader(bytes.NewReader(catalog))

			if err != nil {
				err_ch <- fmt.Errorf("Failed to create CSV reader for catalog (%s), %w", path, err)
				return
			}

			for {
				row, err := catalog_r.Read()

				if err == io.EOF {
					break
				}

				if err != nil {
					err_ch <- fmt.Errorf("Failed to read CSV row for catalog (%s), %w", path, err)
					return
				}

				repo_id := row["repo_id"]
				repo_name := sources_map[repo_id]

				row_ch <- map[string]string{"id": row["id"], "repo": repo_name}
			}

		}(path)

	}

	remaining := len(paths)

	mu := new(sync.RWMutex)

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			log.Fatal(err)
		case row := <-row_ch:
			mu.Lock()
			csv_wr.WriteRow(row)
			mu.Unlock()
		}
	}

}
