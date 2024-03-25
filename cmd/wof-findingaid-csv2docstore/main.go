package main

import ()

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"

	"github.com/aaronland/go-aws-dynamodb"
	"github.com/sfomuseum/go-csvdict"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/producer/docstore"
	gc_docstore "gocloud.dev/docstore"
	gc_dynamodb "gocloud.dev/docstore/awsdynamodb"
)

func main() {

	docstore_uri := flag.String("docstore-uri", "", "...")
	flag.Parse()

	archives := flag.Args()

	ctx := context.Background()

	d := "PT60S"
	monitor, err := timings.NewCounterMonitor(ctx, d)

	if err != nil {
		log.Fatalf("Failed to create timings monitor, %v", err)
	}

	monitor.Start(ctx, os.Stdout)
	defer monitor.Stop(ctx)

	// START OF put me in a package function or something

	var collection *gc_docstore.Collection

	u, err := url.Parse(*docstore_uri)

	if err != nil {
		log.Fatalf("Failed to parse URI, %v", err)
	}

	q := u.Query()

	if u.Scheme == "awsdynamodb" {

		// Connect local dynamodb using Golang
		// https://gist.github.com/Tamal/02776c3e2db7eec73c001225ff52e827
		// https://gocloud.dev/howto/docstore/#dynamodb-ctor

		client, err := dynamodb.NewClientWithURI(ctx, *docstore_uri)

		if err != nil {
			log.Fatalf("Failed to create client, %v", err)
		}

		u, _ := url.Parse(*docstore_uri)
		table_name := u.Host

		partition_key := q.Get("partition_key")

		col, err := gc_dynamodb.OpenCollection(client, table_name, partition_key, "", nil)

		if err != nil {
			log.Fatalf("Failed to open collection for %s (%s), %v", table_name, partition_key, err)
		}

		collection = col

	} else {

		col, err := gc_docstore.OpenCollection(ctx, *docstore_uri)

		if err != nil {
			log.Fatalf("Failed to create collection for '%s', %v", *docstore_uri, err)
		}

		collection = col
	}

	// END OF put me in a package function or something

	for _, path := range archives {

		err := processArchive(ctx, path, collection, monitor)

		if err != nil {
			log.Fatalf("Failed to process %s, %v", path, err)
		}
	}
}

func processArchive(ctx context.Context, path string, collection *gc_docstore.Collection, monitor timings.Monitor) error {

	log.Printf("Process %s\n", path)

	f, err := os.Open(path)

	if err != nil {
		return fmt.Errorf("Failed to open archive %s, %v", path, err)
	}

	defer f.Close()

	return processArchiveWithReader(ctx, f, collection, monitor)
}

func processArchiveWithReader(ctx context.Context, r io.Reader, collection *gc_docstore.Collection, monitor timings.Monitor) error {

	gzip_r, err := gzip.NewReader(r)

	if err != nil {
		return fmt.Errorf("Failed to unzip archive, %w", err)
	}

	tar_r := tar.NewReader(gzip_r)

	sources_tmp := ""
	catalog_tmp := ""

	defer func() {

		to_remove := []string{
			sources_tmp,
			catalog_tmp,
		}

		for _, p := range to_remove {

			if p == "" {
				continue
			}

			os.Remove(p)
		}
	}()

	for {

		header, err := tar_r.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("..., %w", err)
		}

		// Account for 0-length CSV files (repos with no records)

		if header.Size == 0 {
			continue
		}

		switch header.Name {
		case "sources.csv":

			fname, err := writeTempFile(tar_r)

			if err != nil {
				return fmt.Errorf("Failed to create temp file for %s, %w", header.Name, err)
			}

			sources_tmp = fname

		case "catalog.csv":

			fname, err := writeTempFile(tar_r)

			if err != nil {
				return fmt.Errorf("Failed to create temp file for %s, %w", header.Name, err)
			}

			catalog_tmp = fname

		default:
			// pass
		}

		if err != nil {
			return fmt.Errorf("Failed to process %s, %w", header.Name, err)
		}
	}

	if sources_tmp == "" || catalog_tmp == "" {
		return nil
	}

	sources_r, err := os.Open(sources_tmp)

	if err != nil {
		return fmt.Errorf("Failed to open %s, %w", sources_tmp, err)
	}

	defer sources_r.Close()

	lookup, err := processSources(ctx, sources_r)

	if err != nil {
		return fmt.Errorf("Failed to derive sources lookup, %w", err)
	}

	catalog_r, err := os.Open(catalog_tmp)

	if err != nil {
		return fmt.Errorf("Failed to open %s, %w", catalog_tmp, err)
	}

	defer catalog_r.Close()

	err = processCatalog(ctx, catalog_r, lookup, collection, monitor)

	if err != nil {
		return fmt.Errorf("Failed to process catalog, %w", err)
	}

	return nil
}

func processCatalog(ctx context.Context, r io.Reader, lookup map[int64]string, collection *gc_docstore.Collection, monitor timings.Monitor) error {

	csv_r, err := csvdict.NewReader(r)

	if err != nil {
		return fmt.Errorf("Failed to create CSV reader, %w", err)
	}

	for {
		row, err := csv_r.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("Failed to read row, %w", err)
		}

		str_id, ok := row["id"]

		if !ok {
			return fmt.Errorf("Row is missing 'id' column")
		}

		str_repo_id, ok := row["repo_id"]

		if !ok {
			return fmt.Errorf("Row is missing 'repo_id' column")
		}

		id, err := strconv.ParseInt(str_id, 10, 64)

		if err != nil {
			return fmt.Errorf("Failed to parse %s, %w", str_id, err)
		}

		repo_id, err := strconv.ParseInt(str_repo_id, 10, 64)

		if err != nil {
			return fmt.Errorf("Failed to parse %s, %w", str_repo_id, err)
		}

		repo_name, ok := lookup[repo_id]

		if !ok {
			return fmt.Errorf("Missing lookup entry for %d", repo_id)
		}

		err = docstore.AddToCatalog(ctx, collection, id, repo_name)

		if err != nil {
			return fmt.Errorf("Failed to add row to catalog, %w", err)
		}

		go monitor.Signal(ctx)
	}

	return nil
}

func processSources(ctx context.Context, r io.Reader) (map[int64]string, error) {

	lookup := make(map[int64]string)

	csv_r, err := csvdict.NewReader(r)

	if err != nil {
		return nil, fmt.Errorf("Failed to create CSV reader, %w", err)
	}

	for {
		row, err := csv_r.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("Failed to read row, %w", err)
		}

		str_id, ok := row["id"]

		if !ok {
			return nil, fmt.Errorf("Row is missing 'id' column")
		}

		name, ok := row["name"]

		if !ok {
			return nil, fmt.Errorf("Row is missing 'name' column")
		}

		id, err := strconv.ParseInt(str_id, 10, 64)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse %s, %w", str_id, err)
		}

		lookup[id] = name
	}

	return lookup, nil
}

func writeTempFile(r io.Reader) (string, error) {

	wr, err := ioutil.TempFile("", "docstore")

	if err != nil {
		return "", fmt.Errorf("Failed to create temp file, %w", err)
	}

	_, err = io.Copy(wr, r)

	if err != nil {
		return "", fmt.Errorf("Failed to write temp file, %w", err)
	}

	err = wr.Close()

	if err != nil {
		return "", fmt.Errorf("Failed to close temp file, %w", err)
	}

	tmpname := wr.Name()
	return tmpname, nil
}
