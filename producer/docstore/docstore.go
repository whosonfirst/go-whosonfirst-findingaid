package docstore

import (
	"context"
	"fmt"

	gc_docstore "gocloud.dev/docstore"
)

type CatalogRecord struct {
	Id       int64  `docstore:"id"`
	RepoName string `docstore:"repo_name"`
}

func AddToCatalog(ctx context.Context, collection *gc_docstore.Collection, id int64, repo_name string) error {

	test_doc := map[string]interface{}{
		"id":        id,
		"repo_name": "",
	}

	err := collection.Get(ctx, test_doc)

	if err == nil {

		if test_doc["repo_name"] == repo_name {
			return nil
		}
	}

	doc := &CatalogRecord{
		Id:       id,
		RepoName: repo_name,
	}

	err = collection.Put(ctx, doc)

	if err != nil {
		return fmt.Errorf("Failed to put catalog record for %d, %w", id, err)
	}

	return nil
}
