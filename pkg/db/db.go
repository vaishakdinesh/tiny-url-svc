package db

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TODO:: config
func NewDB(ctx context.Context, l *zap.Logger) (*mongo.Client, error) {
	return mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
}
