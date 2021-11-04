package docstore

import (
	"context"
	"fmt"
	gc_docstore "gocloud.dev/docstore"
	"log"
)

type CatalogRecord struct {
	Id       int64  `json:"id"`
	RepoName string `json:"repo_name"`
}

/*

> make cli && ./bin/populate -producer-uri 'awsdynamodb://findinaid?region=us-west-2&endpoint=http://localhost:8000&credentials=static:local:local:local' /usr/local/data/sfomuseum-data-maps/
go build -mod vendor -o bin/populate cmd/populate/main.go
go build -mod vendor -o bin/sources cmd/sources/main.go
go build -mod vendor -o bin/csv2sql cmd/csv2sql/main.go
2021/11/04 12:13:21 ERR missing document key (code=InvalidArgument)
2021/11/04 12:13:21 ERR missing document key (code=InvalidArgument)
2021/11/04 12:13:21 time to index paths (1) 3.778001946s
2021/11/04 12:13:21 ERR missing document key (code=InvalidArgument)
2021/11/04 12:13:21 Failed to populate finding aid, Failed to iterate sources, Failed crawl callback for /usr/local/data/sfomuseum-data-maps/data/171/295/239/3/1712952393.geojson: Failed to store /usr/local/data/sfomuseum-data-maps/data/171/295/239/3/1712952393.geojson, Failed to put catalog record for 1712952393, missing document key (code=InvalidArgument)
2021/11/04 12:13:21 ERR missing document key (code=InvalidArgument)

*/

func AddToCatalog(ctx context.Context, collection *gc_docstore.Collection, id int64, repo_name string) error {

	doc := &CatalogRecord{
		Id:       id,
		RepoName: repo_name,
	}

	err := collection.Put(ctx, doc)

	if err != nil {
		log.Println("ERR", err)
		return fmt.Errorf("Failed to put catalog record for %d, %w", id, err)
	}

	return nil
}
