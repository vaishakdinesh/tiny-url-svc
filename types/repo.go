package types

import "context"

type URLRepo interface {
	Put(ctx context.Context, key string, document URLDocument) error
	Delete(ctx context.Context, key string) error
}
