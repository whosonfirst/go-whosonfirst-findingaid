package main

import (
	_ "github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-reader-http"
	_ "github.com/whosonfirst/go-whosonfirst-findingaid/repo"
	_ "github.com/whosonfirst/go-whosonfirst-iterate/iterator"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-findingaid/application/lookupd"
	"log"
)

func main() {

	ctx := context.Background()

	app, err := lookupd.NewLookupdApplication(ctx)

	if err != nil {
		log.Fatalf("Failed to create lookupd application, %v", err)
	}

	err = app.Run(ctx)

	if err != nil {
		log.Fatalf("Failed to run lookupd application, %v", err)
	}
}
