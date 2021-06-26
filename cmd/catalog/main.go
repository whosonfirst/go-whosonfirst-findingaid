package main

import (
	_ "github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-reader-http"
	_ "github.com/whosonfirst/go-whosonfirst-findingaid/repo"
	_ "github.com/whosonfirst/go-whosonfirst-iterate/iterator"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-findingaid/application/catalog"
	"log"
)

func main() {

	ctx := context.Background()

	app, err := catalog.NewCatalogApplication(ctx)

	if err != nil {
		log.Fatalf("Failed to create catalog application, %v", err)
	}

	err = app.Run(ctx)

	if err != nil {
		log.Fatalf("Failed to run catalog application, %v", err)
	}
}
