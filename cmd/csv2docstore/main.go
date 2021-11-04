package main

import ()

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"github.com/sfomuseum/go-csvdict"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/producer/docstore"
	"io"
	"log"
	"os"
	"strconv"
)

func main() {

	docstore_uri := flag.String("docstore-uri", "", "...")
	flag.Parse()

	archives := flag.Args()

	ctx := context.Background()

	var collection *docstore.Collection

	for _, path := range archives {

		err := processArchive(ctx, path, collection)

		if err != nil {
			log.Fatalf("Failed to process %s, %v", path, err)
		}
	}
}

func processArchive(ctx context.Context, path string, collection *docstore.Collection) error {

	f, err := os.Open(path)

	if err != nil {
		return fmt.Errorf("Failed to open archive %s, %w", path, err)
	}

	defer f.Close()

	return processArchiveWithReader(ctx, f, collection)
}

func processArchiveWithReader(ctx context.Context, r io.Reader, collection *docstore.Collection) error {

	gzip_r, err := gzip.NewReader(r)

	if err != nil {
		return fmt.Errorf("Failed to unzip archive, %w", err)
	}

	tar_r := tar.NewReader(gzip_r)

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
			err = processSources(ctx, tar_r)
		case "catalog.csv":
			err = processCatalog(ctx, tar_r, collection)
		default:
			// pass
		}

		if err != nil {
			return fmt.Errorf("Failed to process %s, %w", header.Name, err)
		}
	}

	return nil
}

func processCatalog(ctx context.Context, r io.Reader, collection *docstore.Collection) error {

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

		repo_name := "FIXME"

		err = docstore.AddToCatalog(ctx, collection, id, repo_id, repo_name)

		if err != nil {
			return fmt.Errorf("Failed to add row to catalog, %w", err)
		}
	}

	return nil
}

func processSources(ctx context.Context, r io.Reader) error {

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

		name, ok := row["name"]

		if !ok {
			return fmt.Errorf("Row is missing 'name' column")
		}

		id, err := strconv.ParseInt(str_id, 10, 64)

		if err != nil {
			return fmt.Errorf("Failed to parse %s, %w", str_id, err)
		}

	}

	return nil
}
