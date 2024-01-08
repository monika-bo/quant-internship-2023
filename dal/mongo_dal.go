package dal

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDal struct {
	client   *mongo.Client
	database *mongo.Database
}

func (m MongoDal) FindByID(ctx context.Context, collection string, id string, document any) error {
	objId, _ := primitive.ObjectIDFromHex(id)
	result := m.database.Collection(collection).
		FindOne(ctx, bson.M{"_id": objId})

	if result.Err() != nil {
		return result.Err()
	}

	if errDecode := result.Decode(document); errDecode != nil {
		return errDecode
	}

	return nil
}

func (m MongoDal) FindAndDeleteByID(ctx context.Context, collection string, id string, document interface{}) error {
	objId, _ := primitive.ObjectIDFromHex(id)
	result := m.database.Collection(collection).
		FindOneAndDelete(ctx, bson.M{"_id": objId})

	if result.Err() != nil {
		return result.Err()
	}

	if errDecode := result.Decode(document); errDecode != nil {
		return errDecode
	}

	return nil
}

func NewMongoDal(ctx context.Context, address string, databaseName string) (*MongoDal, error) {
	clientOptions := options.Client().
		ApplyURI(address)

	client, errConnect := mongo.Connect(ctx, clientOptions)
	if errConnect != nil {
		return nil, errConnect
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return &MongoDal{
		client:   client,
		database: client.Database(databaseName),
	}, nil
}

func (m MongoDal) Update(ctx context.Context, collection string, filter any, update any, optionFuncs ...UpdateOptionsFunc) (*UpdateResult, error) {
	updateOptions := options.Update()

	opts := UpdateOptions{}
	for _, optFunc := range optionFuncs {
		optFunc(&opts)
	}

	if opts.Upsert {
		updateOptions.SetUpsert(true)
	}

	result, err := m.database.
		Collection(collection).
		UpdateOne(ctx, filter, update, updateOptions)

	if err != nil {
		return nil, fmt.Errorf("error while updating document in %s: %w", collection, err)
	}

	return &UpdateResult{
		ModifiedCount: result.ModifiedCount,
		UpsertedID:    result.UpsertedID,
		UpsertedCount: result.UpsertedCount,
		MatchedCount:  result.MatchedCount,
	}, nil
}

func (m MongoDal) Find(ctx context.Context, collection string, findArguments FindArguments, documents any) error {
	findOptions := options.
		Find()

	if findArguments.Projection != nil {
		findOptions.
			SetProjection(findArguments.Projection.ToMongoProjection())
	}

	if findArguments.Sort != nil {
		findOptions.
			SetSort(findArguments.Sort.ToMongoSorting())
	}

	if findArguments.Skip != nil {
		findOptions.
			SetSkip(*findArguments.Skip)
	}

	if findArguments.Limit != nil {
		findOptions.
			SetLimit(*findArguments.Limit)
	}

	var filter interface{}
	filter = bson.M{}
	if findArguments.Filter != nil {
		filter = findArguments.Filter
	}

	cursor, err := m.database.
		Collection(collection).
		Find(ctx, filter, findOptions)

	if err != nil {
		return fmt.Errorf("error finding documents in %s: %w", collection, err)
	}

	return cursor.All(ctx, documents)
}

func (m MongoDal) Disconnect(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

func (m MongoDal) Insert(ctx context.Context, collection string, document []any) (*InsertResult, error) {
	result, err := m.database.
		Collection(collection).
		InsertMany(ctx, document)

	if err != nil {
		return nil, fmt.Errorf("error while inserting documnets in %s: %w", collection, err)
	}

	return &InsertResult{
		InsertedDocumentsCount: len(result.InsertedIDs),
	}, nil
}

func (m MongoDal) Aggregate(ctx context.Context, collection string, pipeline []any, result any) error {
	cursor, err := m.database.
		Collection(collection).
		Aggregate(ctx, pipeline)

	if err != nil {
		return fmt.Errorf("error finding documents in %s: %w", collection, err)
	}

	if err := cursor.All(ctx, result); err != nil {
		return fmt.Errorf("error reading results %w", err)
	}
	return nil

}

func (m MongoDal) UpdateByID(ctx context.Context, collection string, id string, update any, optionFuncs ...UpdateOptionsFunc) (*UpdateResult, error) {
	updateOptions := options.Update()

	opts := UpdateOptions{}
	for _, optFunc := range optionFuncs {
		optFunc(&opts)
	}

	if opts.Upsert {
		updateOptions.SetUpsert(true)
	}

	objID, _ := primitive.ObjectIDFromHex(id)
	result, err := m.database.
		Collection(collection).
		UpdateByID(ctx, objID, update, updateOptions)

	if err != nil {
		return nil, fmt.Errorf("error while updating document in %s: %w", collection, err)
	}

	return &UpdateResult{
		ModifiedCount: result.ModifiedCount,
		UpsertedID:    result.UpsertedID,
		UpsertedCount: result.UpsertedCount,
		MatchedCount:  result.MatchedCount,
	}, nil
}

func (m MongoDal) Delete(ctx context.Context, collection string, filter any) (*DeleteResult, error) {
	result, err := m.database.Collection(collection).DeleteMany(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("Error deleting documents in %s, %w", collection, err)
	}

	return &DeleteResult{
		DeletedCount: result.DeletedCount,
	}, nil
}
