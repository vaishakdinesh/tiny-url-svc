package db

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/vaishakdinesh/tiny-url-svc/types"
)

const (
	dbName         = "tiny-url"
	collectionName = "tiny_urls"
)

type repo struct {
	client *mongo.Client
}

func NewURLRepo(c *mongo.Client) types.URLRepo {
	return &repo{client: c}
}

func (r *repo) Put(ctx context.Context, document any) error {
	collection := r.collection()
	options.InsertOne()
	_, err := collection.InsertOne(ctx, document)
	return err
}

func (r *repo) GetDocument(ctx context.Context, urlKey string) (types.URLDocument, error) {
	urlDoc := &types.URLDocument{}
	collection := r.collection()
	filter := bson.D{{"url_key", urlKey}}
	err := collection.FindOne(ctx, filter).Decode(urlDoc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return types.URLDocument{}, types.ErrDocumentNotFound
		}
		return types.URLDocument{}, err
	}
	return *urlDoc, nil
}

func (r *repo) Delete(ctx context.Context, urlKey string) error {
	collection := r.collection()
	filter := bson.D{{"url_key", urlKey}}
	deleted, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if deleted.DeletedCount == 0 {
		return types.ErrDocumentNotFound
	}
	return err

}

func (r *repo) collection() *mongo.Collection {
	return r.client.Database(dbName).Collection(collectionName)
}
