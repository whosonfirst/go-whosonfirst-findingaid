package findingaid

import (
	"context"
)

type FindingAid interface {
	Index(context.Context, ...string) error
	LookupID(context.Context, int64) (string, error)
}
