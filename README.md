# go-whosonfirst-findingaid

A Go language interface for building and querying finding aids of Who's On First documents.

## Example

```
package main

import (
	_ "github.com/whosonfirst/go-cache"
	_ "github.com/whosonfirst/go-whosonfirst-findingaid/repo"
	_ "github.com/whosonfirst/go-whosonfirst-iterator"	
)

import (
	"context"
	"fmt"

	"github.com/whosonfirst/go-whosonfirst-findingaid"	
	"log"
)

func main(){

	ctx := context.Background()
	
	wof_id := int64(1444838459)

	wof_repo := "whosonfirst-data-admin-is"	
	repo_url := fmt.Sprintf("fixtures/%s", wof_repo)
	
	cache_uri := "gocache://"	// https://github.com/whosonfirst/go-cache
	indexer_uri := "repo://"	// https://github.com/whosonfirst/go-whosonfirst-iterate
	
	fa_uri := fmt.Sprintf("repo://?cache=%s&indexer=%s", cache_uri, indexer_uri)
	
	fa, _ := findingaid.NewFindingAid(ctx, fa_uri)

	fa.Index(ctx, repo_url)

	fa_rsp, _ := fa.LookupID(ctx, wof_id)

	rsp := fa_rsp.(*repo.FindingAidResponse)
	
	if rsp.Repo != wof_repo {
		log.Fatal("Invalid repo")
	}
}
```

_Error handling omitted for the sake of brevity._

## FindingAids

Conceptually a finding aid consists of two parts:

* An indexer which indexes (or catalogs) one or more Who's On First (WOF) records in to a cache. WOF records may be cataloged in full, truncated or otherwise manipulated according to logic implemented by the indexing or caching layers.
* A cache of WOF records, in full or otherwise manipulated, that can resolved using a given WOF ID.

It is generally assumed that a complete catalog of WOF records will be assembled in advance of any query actions but that is not an absolute requirement. For an example of a lazy-loading catalog and query implementation, where all operations are performed at runtime, consult the documentation for the `readercache` chaching layer below.

There can be more than one kind of finding aid. Finding aids can implement their own internal logic for cataloging, caching and querying WOF records. A finding aid need only implement the following interface:

```
type FindingAid interface {
	Index(context.Context, ...string) error
	IndexReader(context.Context, io.Reader) error
	LookupID(context.Context, int64) (interface{}, error)
}
```

Note the ambiguous return value (`interface{}`) for the `LookupID` method. Since it impossible to know in advance the response properties of any given finding aid it is left to developers to cast query results in to the appropriate type if necessary.

The `findingaid` package provides for a generic constructor method using URI strings to distinguish one finding from another. For example:

```
import (
       "context"
	_ "github.com/whosonfirst/go-whosonfirst-findingaid/repo"
)

func main() {
	ctx := context.Background()

	fa_uri := "repo://?cache={cache_uri}&indexer={indexer_uri}"
	fa, _ := findingaid.NewFindingAid(ctx, fa_uri)
}	
```

Individual finding aid implementations must "register" themselves and their URI schemes on initialization in order to make themselves available. For example:

```
package repo

import (
       "context"
	"github.com/whosonfirst/go-whosonfirst-findingaid"
)

func init() {

	ctx := context.Background()
	err := findingaid.RegisterFindingAid(ctx, "repo", NewRepoFindingAid)

	if err != nil {
		panic(err)
	}
}

func NewRepoFindingAid(ctx context.Context, uri string) (findingaid.FindingAid, error) {
	...
}	
```

The following finding aids are available by default:

### repo

This finding aid catalogs Who's On First records and the `wof:repo` name. For details on why this is necessary consult the [Upcoming changes to Who's On First administrative data](https://whosonfirst.org/blog/2019/05/09/changes/) blog post.

The return value of this finding aid's `LookupID` method is a pointer to a `repo.FindingAidResponse` struct:

```
type FindingAidResponse struct {
	ID   int64  `json:"id"`
	Repo string `json:"repo"`
	URI  string `json:"uri"`
}
```

To load this finding aid you would add the following (abbreviated code) to your code:

```
import (
       "context"
	_ "github.com/whosonfirst/go-whosonfirst-findingaid/repo"
)

func main() {
	ctx := context.Background()

	fa_uri := "repo://?cache={cache_uri}&indexer={indexer_uri}"
	fa, _ := findingaid.NewFindingAid(ctx, fa_uri)
}	
```

URI strings for the `repo` finding aid take the following parameters:

| Name | Type | Rquired | Notes |
| --- | --- | --- | --- |
| cache | string | yes | A valid `whosonfirst/go-cache URI string. |
| indexer | string | yes | A valid `whosonfirst/go-whosonfirst-iterate/emitter URI string. |

## Caches

The `go-whosonfirst-findingaid` package imports the [whosonfirst/go-cache](https://github.com/whosonfirst/go-cache) package so all the caches it exports are automatically available. Please consult [that package's documentation](#) for details. The following additional caching layers are also available:

### readercache

The `readercache` package implements the `whosonfirst/go-cache` interface by lazy-loading cache values using a valid [whosonfirst/go-reader](https://github.com/whosonfirst/go-reader) `Reader` instance. 

For example, this package is used in concert with the `null` indexing package by the [lookupd](cmd/lookupd) tool to implement an HTTP findingaid that resolves, and caches, indentifiers at runtime.

```
import (
	_ "github.com/whosonfirst/go-whosonfirst-findingaid/repo"
)

cache_uri := "readercache://?reader={READER_URI}&cache={CACHE_URI}"
```

Where `{READER_URI}` is a valid [whosonfirst/go-reader](https://github.com/whosonfirst/go-reader) URI string and `{CACHE_URI}` is a valid [whosonfirst/go-cache](https://github.com/whosonfirst/go-cache) URI string.

## Indexers

The `go-whosonfirst-findingaid` package imports the [whosonfirst/go-whosonfirst-iterate](https://github.com/whosonfirst/go-whosonfirst-iterate) package so all the iterators and emitters it exports are automatically available. Please consult [that package's documentation](https://github.com/whosonfirst/go-whosonfirst-iterate) for details.

## Tools

### lookupd

A simple HTTP server for querying a finding aid by URI and returning the results as JSON.

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

As a convenience the string values of `{cache_uri}` and `{indexer_uri}` in the value of the `-findingaid-uri` argument will be automatically replaced with the values of the `-cache-uri` and `-indexer-uri` flags. Values will also be automatically URL encoded so you don't need to do that yourself.

For example:

```
$> go run -mod vendor cmd/lookupd/main.go
2020/10/12 13:47:24 Listening on http://localhost:8080
```

And then:

```
$> curl -s localhost:8080/85922583 | jq
{
  "id": 85922583,
  "repo": "whosonfirst-data-admin-us",
  "uri": "859/225/83/85922583.geojson"
}
```

#### AWS Lambda

...

## See also

* https://github.com/whosonfirst/go-cache
* https://github.com/whosonfirst/go-whosonfirst-iterate
* https://github.com/whosonfirst/go-reader
* https://en.wikipedia.org/wiki/Finding_aid