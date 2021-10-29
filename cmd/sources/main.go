package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/provider"
	"log"
)

func main() {

	provider_uri := flag.String("provider-uri", "github://whosonfirst-data", "...")

	uri_template := flag.String("uri-template", "", "...")

	flag.Parse()

	ctx := context.Background()

	pr, err := provider.NewProvider(ctx, *provider_uri)

	if err != nil {
		log.Fatalf("Failed to create new provider, %v", err)
	}

	var sources []string

	if *uri_template != "" {
		sources, err = pr.IteratorSourcesWithURITemplate(ctx, *uri_template)
	} else {
		sources, err = pr.IteratorSources(ctx)
	}

	if err != nil {
		log.Fatalf("Failed to derive sources, %v", err)
	}

	for _, s := range sources {
		fmt.Println(s)
	}
}
