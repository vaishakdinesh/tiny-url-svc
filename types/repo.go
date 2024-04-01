package types

import "context"

// URLRepo abstraction for the repository to store tiny urls
type URLRepo interface {
	Put(ctx context.Context, document any) error
	GetDocument(ctx context.Context, urlKey string) (URLDocument, error)
	Delete(ctx context.Context, urlKey string) error
}
