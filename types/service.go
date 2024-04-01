package types

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	urlFormat = "%s://%s/tinyurlsvc/%s"
)

type Metrics interface {
	RegisterProm() error
}

// URLService represents domain service abstraction where biz logic resides.
type URLService interface {
	Metrics
	GenerateTinyURL(ctx context.Context, longUrl string, liveForever bool) (URLDocument, error)
	GetTinyURL(ctx context.Context, urlKey string) (URLDocument, error)
	DeleteTinyURL(ctx context.Context, urlKey string) error
}

type URLDocument struct {
	Base10ID    int64     `bson:"base_10_id"`
	URLKey      string    `bson:"url_key"`
	LongURL     string    `bson:"long_url"`
	ExpireTime  time.Time `bson:"expire_time"`
	LiveForever bool      `bson:"live_forever"`
}

func (u URLDocument) ToURL(ctx echo.Context) string {
	scheme := "http"
	if ctx.Request().TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf(urlFormat, scheme, ctx.Request().Host, u.URLKey)
}

type CacheService interface {
	Cache(ctx context.Context, key string, val any) error
	Delete(ctx context.Context, key string) error
	GetCachedValue(ctx context.Context, key string) (string, error)
}
