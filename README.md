# go-whosonfirst-findingaid

Work in progress.

## Example

```
package main

import (
	"context"
	"fmt"
	_ "github.com/whosonfirst/go-cache"			
	"github.com/whosonfirst/go-whosonfirst-findingaid"	
	_ "github.com/whosonfirst/go-whosonfirst-findingaid/repo"
	_ "github.com/whosonfirst/go-whosonfirst-index/fs"
	"log"
)

func main(){

	ctx := context.Background()
	
	wof_id := int64(1444838459)

	wof_repo := "whosonfirst-data-admin-is"	
	repo_url := fmt.Sprintf("fixtures/%s", wof_repo)
	
	cache_uri := "gocache://"	// https://github.com/whosonfirst/go-cache
	indexer_uri := "repo://"	// https://github.com/whosonfirst/go-whosonfirst-index
	
	fa_uri := fmt.Sprintf("repo://?cache=%s&indexer=%s", cache_uri, indexer_uri)
	
	fa, _ := findingaid.NewFindingAid(ctx, fa_uri)

	fa.Index(ctx, repo_url)

	rsp, _ := fa.LookupID(ctx, wof_id)

	if rsp.Repo != wof_repo {
		log.Fatal("Invalid repo")
	}
}
```

_Error handling omitted for the sake of brevity._

## Tools

### lookupd

```
$> ./bin/lookupd -h
  -cache-uri string
    	A valid whosonfirst/go-cache URI string. (default "readercache://?reader=http://data.whosonfirst.org&cache=gocache://")
  -enable-cors
    	Enable CORS headers for output. (default true)
  -findingaid-uri string
    	A valid whosonfirst/go-whosonfirst-findingaid URI string. (default "repo://?cache={cache_uri}&indexer={indexer_uri}")
  -indexer-uri string
    	A valid whosonfirst/go-whosonfirst-index URI string. (default "null://")
  -server-uri string
    	A valid aaronland/go-http-server URI string. (default "http://localhost:8080")
```


## FindingAids

```
type FindingAid interface {
	Index(context.Context, ...string) error
	IndexReader(context.Context, io.Reader) error
	LookupID(context.Context, int64) (interface{}, error)
}
```

### Repo(sitory) FindingAid

## Caches

This package imports the [whosonfirst/go-cache](#) package so all the caches it exports are automatically available. Please consult [that package's documentation](#) for details. The following additional caching layers are also available:

### readercache

The `readercache` package implements the `whosonfirst/go-cache` interface by lazy-loading cache values using a valid [whosonfirst/go-reader](#) `Reader` instance. 

For example, this package is used in concert with the `null` indexing package by the [lookupd](cmd/lookupd) tool to implement an HTTP findingaid that resolves, and caches, indentifiers at runtime.

```
import (
	_ "github.com/whosonfirst/go-whosonfirst-findingaid/repo"
)

cache_uri := "readercache://?reader={READER_URI}&cache={CACHE_URI}"
```

Where `{READER_URI}` is a valid [whosonfirst/go-reader](#) URI string and `{CACHE_URI}` is a valid [whosonfirst/go-cache](#) URI string.

## Indexers

This package imports the [whosonfirst/go-whosonfirst-index](#) package so all the caches it exports are automatically available. Please consult [that package's documentation](#) for details. The following additional caching layers are also available:

### null

The `null` package implements to `whosonfirst/go-whosonfirst-index` interface but doesn't actually index anything at all.

For example, this package is used in concert with the `readercache` caching package by the [lookupd](cmd/lookupd) tool to implement an HTTP findingaid that resolves, and caches, indentifiers at runtime.

```
import (
	_ "github.com/whosonfirst/go-whosonfirst-findingaid/index"
)

indexer_uri := "null://"
```

## See also

* https://github.com/whosonfirst/go-cache
* https://github.com/whosonfirst/go-whosonfirst-index
* https://en.wikipedia.org/wiki/Finding_aid