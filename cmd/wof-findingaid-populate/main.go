package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v2"
	_ "gocloud.dev/docstore/awsdynamodb"
	_ "gocloud.dev/docstore/memdocstore"

	"github.com/jtacoma/uritemplates"
	"github.com/sfomuseum/go-timings"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/producer"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/provider"
)

func main() {

	iterator_uri := flag.String("iterator-uri", "repo://", "A valid whosonfirst/go-whosonfirst-iterate/v2 URI.")

	provider_uri := flag.String("provider-uri", "", "An optional whosonfirst/go-whosonfirst-findingaid/v2/provider URI to use for deriving additional sources.")

	producer_uri := flag.String("producer-uri", "csv://?archive=archive.tar.gz", "A valid whosonfirst/go-whosonfirst-findingaid/v2/producer URI.")

	atomic := flag.Bool("atomic", false, "Produce atomic findingaids for each item in a source list. If true then -producer URI must be a valid URI template containing a '{source}' variable to expand with findingaid name.")

	flag.Parse()

	ctx := context.Background()

	iterator_sources := flag.Args()

	var prd producer.Producer
	var uri_t *uritemplates.UriTemplate

	if *atomic {

		t, err := uritemplates.Parse(*producer_uri)

		if err != nil {
			log.Fatalf("Unable to parse -producer-uri flag as a URI template, %v", err)
		}

		str_names := strings.Join(t.Names(), "")

		if str_names != "source" {
			log.Fatalf("Unexpected URI template, must only contain a {source} variable")
		}

		p, err := producer.NewProducer(ctx, "null://")

		if err != nil {
			log.Fatalf("Failed to create null producer, %v", err)
		}

		prd = p
		uri_t = t

	} else {

		p, err := producer.NewProducer(ctx, *producer_uri)

		if err != nil {
			log.Fatalf("Failed to create new producer, %v", err)
		}

		prd = p
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

	monitor, err := timings.NewMonitor(ctx, "counter://PT60S")

	if err != nil {
		log.Fatalf("Failed to create timings monitor, %v", err)
	}

	monitor.Start(ctx, os.Stdout)
	defer monitor.Stop(ctx)

	if *atomic {

		for _, src := range iterator_sources {

			n := filepath.Base(src)

			values := map[string]interface{}{
				"source": n,
			}

			local_uri, err := uri_t.Expand(values)

			if err != nil {
				log.Fatalf("Failed to expand URI template for %s, %v", src, err)
			}

			prd, err := producer.NewProducer(ctx, local_uri)

			if err != nil {
				log.Fatalf("Failed to create new producer for %s, %v", local_uri, err)
			}

			err = prd.PopulateWithIterator(ctx, monitor, *iterator_uri, src)

			if err != nil {
				log.Fatalf("Failed to populate finding aid for %s, %v", src, err)
			}

			err = prd.Close(ctx)

			if err != nil {
				log.Fatalf("Failed to close producer for %s, %v", src, err)
			}
		}

	} else {

		err = prd.PopulateWithIterator(ctx, monitor, *iterator_uri, iterator_sources...)

		if err != nil {
			log.Fatalf("Failed to populate finding aid, %v", err)
		}
	}
}
