package main

import (
	"context"
	"log"

	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/app/resolve"
)

func main() {

	ctx := context.Background()
	err := resolve.Run(ctx)

	if err != nil {
		log.Fatalf("Failed to run resolve tool, %v", err)
	}
}
