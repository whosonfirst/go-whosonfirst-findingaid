# go-whosonfirst-index-git

Git support for the go-whosonfirst-index index.Indexer interface.

## Example

```
package main

import (
	"context"
	"github.com/whosonfirst/go-whosonfirst-index"
	_ "github.com/whosonfirst/go-whosonfirst-index-git"
	"io"
	"log"
	"sync/atomic"
)

func main() {

	var count int64
	count = 0

	cb := func(ctx context.Context, fh io.Reader, args ...interface{}) error {
		atomic.AddInt64(&count, 1)
		return nil
	}

	i, _ := index.NewIndexer("git://", cb)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	i.Index(ctx, "git@github.com:whosonfirst-data/whosonfirst-data-admin-is.git")
	log.Println(count, i.Indexed)
}
```

_Error handling omitted for the sake of brevity._

## Tools

### wof-index-count

By default `go-whosonfirst-index-git` clones Git repositories in to memory:

```
go run -mod vendor cmd/wof-index-count/main.go -dsn 'git://' git@github.com:whosonfirst-data/whosonfirst-data-admin-is.git
2019/12/17 12:02:40 436 0
```

If your `go-whosonfirst-index-git` URI string (DSN) contains a path then repositories will be cloned in that path:

```
go run -mod vendor cmd/wof-index-count/main.go -dsn 'git:///tmp/data' git@github.com:whosonfirst-data/whosonfirst-data-admin-is.git
2019/12/17 12:02:40 436 0
```

By default repositories cloned in to a path are removed. If you want to preserve the cloned repository include a `?preserve=1` query parameter in your URI string:

```
go run -mod vendor cmd/wof-index-count/main.go -dsn 'git:///tmp/data?preserve=1' git@github.com:whosonfirst-data/whosonfirst-data-admin-is.git
2019/12/17 12:02:40 436 0
```

In this example the clone repository will be store in `/tmp/data/whosonfirst-data-admin-is.git`.

## See also

* https://godoc.org/gopkg.in/src-d/go-git.v4
* https://github.com/whosonfirst/go-whosonfirst-index