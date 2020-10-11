package index

import (
	"context"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
)

func init() {
	ni := NewNullIndexer()
	wof_index.Register("null", ni)
}

type NullIndexer struct {
	wof_index.Driver
}

func NewNullIndexer() wof_index.Driver {
	return &NullIndexer{}
}

func (i *NullIndexer) Open(string) error {
	return nil
}

func IndexURI(context.Context, wof_index.IndexerFunc, string) error {
	return nil
}
