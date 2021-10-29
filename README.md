# go-whosonfirst-findingaid

## Documentation

## Motivation

## Example

### Library

_Error handling omitted for the sake of brevity._

```
package main

import (
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/whosonfirst/go-whosonfirst-iterate-git/v2"
)

import (
	"context"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/producer"
	"github.com/whosonfirst/go-whosonfirst-findingaid/v2/provider"
)

func main() {

	iterator_uri := flag.String("iterator-uri", "git:///tmp", "A valid whosonfirst/go-whosonfirst-iterate/v2 URI.")
	provider_uri := flag.String("provider-uri", "github://whosonfirst-data", "...")
	producer_uri := flag.String("producer-uri", "csv://?archive.tar.gz", "...")

	flag.Parse()

	ctx := context.Background()

	prd, _ := producer.NewProducer(ctx, *producer_uri)

	defer prd.Close(ctx)

	prv, _ := provider.NewProvider(ctx, *provider_uri)

	iterator_sources, _ := prv.IteratorSources(ctx)

	prd.PopulateWithIterator(ctx, *iterator_uri, iterator_sources...)
}

```

### Command line

```
#!/bin/sh

SOURCES=`bin/sources -provider-uri "github://whosonfirst-data?prefix=whosonfirst-data-admin-"`

for REPO in ${SOURCES}
do
    NAME=`basename ${REPO} | sed 's/\.git//g'`
    time bin/populate-sql -iterator-uri git:///tmp -provider-uri ${PROVIDER_URI} -producer-uri "sql://sqlite3/?dsn=/usr/local/data/findingaid/${NAME}.db" ${REPO}
done
```

```
$> ./bin/populate -iterator-uri git:///tmp -provider-uri 'github://sfomuseum-data?prefix=sfomuseum-data-maps'
2021/10/28 20:08:55 time to index paths (1) 2.408854633s

$> tar -tf archive.tar.gz 
catalog.csv
sources.csv
```

## Tools

### populate

```
$> ./bin/populate \
	-iterator-uri git:///tmp \
	-provider-uri 'github://sfomuseum-data?prefix=sfomuseum-data-&exclude=sfomuseum-data-flights&exclude=sfomuseum-data-faa&exclude=sfomuseum-data-garages&exclude=sfomuseum-data-checkpoints' \
	-producer-uri 'csv://?archive=archive.tar.gz'

```

### sources

_TBW_

## See also

* https://www.github.com/whosonfirst/go-whosonfirst-iterate