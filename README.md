# go-whosonfirst-findingaid

Work in progress.

## Example

_Error handling omitted for the sake of brevity._

```
package main

import (
	"context"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-findingaid/repo"
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
	
	fa, _ := repo.NewRepoFindingAid(ctx, fa_uri)

	fa.Index(ctx, repo_url)

	var rsp repo.FindingAidResponse
	
	fa.LookupID(ctx, wof_id, &rsp)

	if rsp.Repo != wof_repo {
		log.Fatal("Invalid repo")
	}
}
```

Notes:

* Eventually there will be a `findingaid.NewFindingAid` helper method.
* The use of `whosonfirst/go-whosonfirst-index` packages will likely be replaced the [whosonfirst/go-whosonfirst-iterate](https://github.com/whosonfirst/go-whosonfirst-iterate) packages but that's still work in progress.

## Interfaces

### FindingAid

```
type FindingAid interface {
	Index(context.Context, ...string) error
	IndexReader(context.Context, io.Reader) error
	LookupID(context.Context, int64, interface{}) error
}
```

## See also

* https://github.com/whosonfirst/go-cache
* https://github.com/whosonfirst/go-whosonfirst-index
* https://en.wikipedia.org/wiki/Finding_aid