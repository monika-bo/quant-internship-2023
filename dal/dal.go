package dal

import (
	"context"
)

// database names
const (
	TestDbName = "test-gif-manager"
	DbName     = "gif-manager"
)

// collection names
const (
	CollCategories = "categories"
	CollGifs       = "gifs"
	CollGroups     = "groups"
	CollUsers      = "users"
)

type DAL interface {
	Disconnect(ctx context.Context) error
	Insert(ctx context.Context, collection string, document []any) (*InsertResult, error)
	Find(ctx context.Context, collection string, findArguments FindArguments, result any) error
	FindByID(ctx context.Context, collection string, id string, result any) error
	Delete(ctx context.Context, collection string, filter any) (*DeleteResult, error)
	FindAndDeleteByID(ctx context.Context, collection string, id string, document interface{}) error
	Aggregate(ctx context.Context, collection string, pipeline []any, result any) error
	Update(ctx context.Context, collection string, filter any, update any, optionFuncs ...UpdateOptionsFunc) (*UpdateResult, error)
	UpdateByID(ctx context.Context, collection string, id string, update any, optionFuncs ...UpdateOptionsFunc) (*UpdateResult, error)
}
