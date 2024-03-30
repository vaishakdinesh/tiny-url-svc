package types

import (
	"context"
	"time"
)

// URLService represents domain service abstraction
// where biz logic resides. For simplicity, we've just embedded
// the repo
type URLService interface {
	GenerateTinyURL(ctx context.Context, longUrl string) (URLDocument, error)
}

type URLDocument struct {
	ID         int64
	TinyURL    string
	URL        string
	ExpireTime time.Time
}
