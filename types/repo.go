package types

import "context"

type URLRepo interface {
	Put(ctx context.Context, document any) error
	GetDocument(ctx context.Context, urlKey string) (URLDocument, error)
	Delete(ctx context.Context, urlKey string) error
}
