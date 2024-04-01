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

// URLService represents domain service abstraction where biz logic resides.
type URLService interface {
	GenerateTinyURL(ctx context.Context, longUrl string) (URLDocument, error)
	GetTinyURL(ctx context.Context, urlKey string) (URLDocument, error)
	DeleteTinyURL(ctx context.Context, urlKey string) error
}

type URLDocument struct {
	Base10ID   int64     `bson:"base_10_id" json:"Base10ID"`
	URLKey     string    `bson:"url_key" json:"URLKey"`
	LongURL    string    `bson:"long_url" json:"LongURL"`
	ExpireTime time.Time `bson:"expire_time" json:"ExpireTime"`
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
