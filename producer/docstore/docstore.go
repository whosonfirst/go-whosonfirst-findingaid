package docstore

import (
	"context"
	"fmt"
	gc_docstore "gocloud.dev/docstore"
)

type CatalogRecord struct {
	Id       int64
	RepoName string
}

func AddToCatalog(ctx context.Context, collection *gc_docstore.Collection, id int64, repo_name string) error {

	doc := CatalogRecord{
		Id:       id,
		RepoName: repo_name,
	}

	err := collection.Put(ctx, doc)

	if err != nil {
		return fmt.Errorf("Failed to put catalog record for %d, %w", id, err)
	}

	return nil
}