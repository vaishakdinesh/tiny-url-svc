package db

import (
	"context"
	"github.com/vaishakdinesh/tiny-url-svc/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type repo struct {
	client *mongo.Client
}

func NewURLRepo(c *mongo.Client) types.URLRepo {
	return &repo{client: c}
}

func (r *repo) Put(ctx context.Context, key string, document types.URLDocument) error {
	collection := r.client.Database("tiny-url").Collection("urls")
	_, err := collection.InsertOne(ctx, bson.D{
		{key, document},
	})
	return err
}

func (r *repo) Delete(ctx context.Context, key string) error {
	collection := r.client.Database("tiny-url").Collection("urls")
	_, err := collection.DeleteOne(ctx, key)
	return err

}
