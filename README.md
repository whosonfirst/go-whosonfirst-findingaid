# go-whosonfirst-findingaid

## Documentation

Documentation is incomplete and will be updated shortly.

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

For a working example have a look at [cmd/populate](cmd/populate/main.go).

### Command line

#### CSV archives

```
$> ./bin/populate -iterator-uri git:///tmp -provider-uri 'github://sfomuseum-data?prefix=sfomuseum-data-maps'
2021/10/28 20:08:55 time to index paths (1) 2.408854633s

$> tar -tf archive.tar.gz 
catalog.csv
sources.csv
```

#### SQLite databases

```
#!/bin/sh

SOURCES=`bin/sources -provider-uri "github://whosonfirst-data?prefix=whosonfirst-data-admin-"`

for REPO in ${SOURCES}
do
    NAME=`basename ${REPO} | sed 's/\.git//g'`
    time bin/populate-sql -iterator-uri git:///tmp -provider-uri ${PROVIDER_URI} -producer-uri "sql://sqlite3/?dsn=/usr/local/data/findingaid/${NAME}.db" ${REPO}
done
```

#### Protobuffers

```
$> ./bin/populate \
	-producer-uri protobuf:///usr/local/data/whosonfirst-data-admin-xy.pb \
	/usr/local/data/whosonfirst-data-admin-xy

$> ll /usr/local/data/whosonfirst-data-admin-xy.pb 
-rw-r--r--  1 wof  wheel  245798 Oct 28 17:13 /usr/local/data/whosonfirst-data-admin-xy.pb
```

## Concepts

### Iterators

An iterator is a valid `whosonfirst/go-whosonfirst-iterate/v2` instance (or URI used to create that instance) that is the source of records to pass to a (findingaid) producer.

### Producers

Producers implement the `producer.Producer` interface and are used to populate finding aids where "populate" means updating a data store with information mapping a Who's On First ID to its corresponding repository name.

### Providers

Providers implement the `provider.Provider` interface and are used to generate a list of iterator URIs for crawling by a producer.

## Tools

### csv2sql

```
$> du -h -d 1 /usr/local/data/findingaid/csv/
15M     /usr/local/data/findingaid/csv/

$> time ./bin/csv2sql -database-uri 'sql://sqlite3?dsn=admin.db' /usr/local/data/findingaid/csv/*.gz

real	1m49.170s
user	1m31.838s
sys	0m22.015s

$> sqlite3 admin.db 
SQLite version 3.7.17 2013-05-20 00:56:22
Enter ".help" for instructions
Enter SQL statements terminated with a ";"
sqlite> SELECT COUNT(id) FROM catalog;
4930544

$> du -h admin.db 
81M   admin.db
```

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