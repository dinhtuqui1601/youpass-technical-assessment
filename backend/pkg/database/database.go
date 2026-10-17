package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect opens a connection to MongoDB and verifies it with a ping.
func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}

// EnsureIndexes creates the indexes required by the application. Index creation
// is idempotent and safe to run on every startup.
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	type indexSpec struct {
		collection string
		models     []mongo.IndexModel
	}

	specs := []indexSpec{
		{
			collection: "users",
			models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
			},
		},
		{
			collection: "auth_sessions",
			models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "sessionId", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "userId", Value: 1}}},
				{Keys: bson.D{{Key: "refreshTokenHash", Value: 1}}},
			},
		},
		{
			collection: "submissions",
			models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "ownerId", Value: 1}}},
			},
		},
		{
			collection: "share_links",
			models: []mongo.IndexModel{
				{Keys: bson.D{{Key: "code", Value: 1}}, Options: options.Index().SetUnique(true)},
				{Keys: bson.D{{Key: "ownerId", Value: 1}}},
				{Keys: bson.D{{Key: "submissionId", Value: 1}}},
			},
		},
		{
			collection: "share_views",
			models: []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: "shareLinkId", Value: 1},
						{Key: "viewId", Value: 1},
					},
					Options: options.Index().SetUnique(true),
				},
				{Keys: bson.D{{Key: "shareLinkId", Value: 1}, {Key: "viewedAt", Value: -1}}},
			},
		},
	}

	for _, spec := range specs {
		if _, err := db.Collection(spec.collection).Indexes().CreateMany(ctx, spec.models); err != nil {
			return err
		}
	}

	return nil
}
