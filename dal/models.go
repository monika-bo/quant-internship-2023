package dal

import (
	"go.mongodb.org/mongo-driver/bson"
)

type InsertResult struct {
	InsertedDocumentsCount int
}

type Projection struct {
	FieldName     string
	ShouldExclude bool
}

type DeleteResult struct {
	DeletedCount int64
}

type Projections []Projection

func (projections Projections) ToMongoProjection() bson.M {
	mongoProjection := bson.M{}
	for _, projection := range projections {
		if projection.ShouldExclude {
			mongoProjection[projection.FieldName] = 0
		} else {
			mongoProjection[projection.FieldName] = 1
		}
	}
	return mongoProjection
}

type Sort struct {
	FieldName string
	Ascending bool
}

type Sorts []Sort

func (sorts Sorts) ToMongoSorting() bson.M {
	mongoSorting := bson.M{}
	for _, sort := range sorts {
		if sort.Ascending {
			mongoSorting[sort.FieldName] = 1
		} else {
			mongoSorting[sort.FieldName] = -1
		}
	}
	return mongoSorting
}

type FindArguments struct {
	Filter     interface{}
	Projection Projections
	Sort       Sorts
	Skip       *int64
	Limit      *int64
}

func NewFindArguments() *FindArguments {
	return &FindArguments{}
}

func (args *FindArguments) WithFilter(filter interface{}) *FindArguments {
	args.Filter = filter
	return args
}

func (args *FindArguments) WithProjection(projection Projections) *FindArguments {
	args.Projection = projection
	return args
}

func (args *FindArguments) WithSorts(sort Sorts) *FindArguments {
	args.Sort = sort
	return args
}

func (args *FindArguments) WithSkip(skip int) *FindArguments {
	skip64 := int64(skip)
	args.Skip = &skip64
	return args
}

func (args *FindArguments) WithLimit(limit int) *FindArguments {
	limit64 := int64(limit)
	args.Limit = &limit64
	return args
}

type UpdateResult struct {
	ModifiedCount int64
	UpsertedCount int64
	UpsertedID    interface{}
	MatchedCount  int64
}
