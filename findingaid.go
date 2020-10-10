package findingaid

import (
	"context"
	"io"
)

type FindingAid interface {
	Index(context.Context, ...string) error
	IndexReader(context.Context, io.Reader) error
	LookupID(context.Context, int64, interface{}) error
}

type FindingAidResponse struct {
	Id   int64  `json:"id"`
	URI  string `json:"uri"`
	Repo string `json:"repo"`
}
