# go-whosonfirst-findingaid

Work in progress.

## Example

_Error handling omitted for the sake of brevity._

```
package main

import (
	"context"
	"fmt"
	_ "github.com/whosonfirst/go-whosonfirst-index/fs"	
	"log"
)

func main(){

	ctx := context.Background()
	
	wof_id := int64(1444838459)

	repo := "whosonfirst-data-admin-is"	
	repo_url := fmt.Sprintf("fixtures/%s", repo)
	
	cache_uri := "gocache://"	// https://github.com/whosonfirst/go-cache
	indexer_uri := "repo://"	// https://github.com/whosonfirst/go-whosonfirst-index
	
	fa_uri := fmt.Sprintf("repo:///?cache=%s&indexer=%s", cache_uri, indexer_uri)
	
	fa, _ := NewRepoFindingAid(ctx, fa_uri)

	fa.Index(ctx, repo_url)

	var rsp FindingAidResponse
	
	fa.LookupID(ctx, wof_id, &rsp)

	if rsp.Repo != repo {
		log.Fatal("Invalid repo")
	}
}
```

## See also

* https://github.com/whosonfirst/go-cache
* https://github.com/whosonfirst/go-whosonfirst-index
* https://en.wikipedia.org/wiki/Finding_aid