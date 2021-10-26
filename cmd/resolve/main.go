package main

import (
	_ "gocloud.dev/blob/fileblob"
)

import (
	_ "github.com/whosonfirst/go-cache-blob"
	_ "github.com/whosonfirst/go-whosonfirst-findingaid/repo"
	_ "github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
)

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-findingaid/application/resolve"
	"log"
)

func main() {

	ctx := context.Background()

	app, err := resolve.NewResolveApplication(ctx)

	if err != nil {
		log.Fatalf("Failed to create resolve application, %v", err)
	}

	err = app.Run(ctx)

	if err != nil {
		log.Fatalf("Failed to run resolve application, %v", err)
	}
}
