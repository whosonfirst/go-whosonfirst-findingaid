package main

import (
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v2"
	_ "gocloud.dev/docstore/awsdynamodb"
	_ "gocloud.dev/docstore/memdocstore"
)

import (
	"context"
	"flag"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/producer"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/provider"
	"log"
	"os"
	"time"
)

func main() {

	iterator_uri := flag.String("iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/v2 URI.")

	provider_uri := flag.String("provider-uri", "", "An optional whosonfirst/go-whosonfirst-findingaid/v2/provider URI to use for deriving additional sources.")

	producer_uri := flag.String("producer-uri", "csv://?archive=archive.tar.gz", "A valid whosonfirst/go-whosonfirst-findingaid/v2/producer URI.")

	flag.Parse()

	ctx := context.Background()

	iterator_sources := flag.Args()

	prd, err := producer.NewProducer(ctx, *producer_uri)

	if err != nil {
		log.Fatalf("Failed to create new producer, %v", err)
	}

	defer prd.Close(ctx)

	if *provider_uri != "" {

		prv, err := provider.NewProvider(ctx, *provider_uri)

		if err != nil {
			log.Fatalf("Failed to create new provider, %v", err)
		}

		sources, err := prv.IteratorSources(ctx)

		if err != nil {
			log.Fatalf("Failed to derive sources, %v", err)
		}

		for _, s := range sources {

			iterator_sources = append(iterator_sources, s)
		}

	}

	d := time.Second * 60
	monitor, err := timings.NewCounterMonitor(ctx, d)

	if err != nil {
		log.Fatalf("Failed to create timings monitor, %v", err)
	}

	monitor.Start(ctx, os.Stdout)
	defer monitor.Stop(ctx)

	err = prd.PopulateWithIterator(ctx, monitor, *iterator_uri, iterator_sources...)

	if err != nil {
		log.Fatalf("Failed to populate finding aid, %v", err)
	}
}
