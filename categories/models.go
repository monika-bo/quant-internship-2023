package categories

import (
	"gifmanager-backend/gifs"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Category struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `bson:"name"`
	UserId   primitive.ObjectID `bson:"userId"`
	GifCount int                `bson:"gifCount"`
}

type Categories []Category

func (c Category) ToDto() CategoryDto {
	return CategoryDto{
		ID:        c.ID.Hex(),
		Name:      c.Name,
		GifsCount: c.GifCount,
	}
}

type GifsByCategory struct {
	CategoryId primitive.ObjectID `bson:"categoryId,omitempty"`
	Name       string             `bson:"name"`
	Gifs       gifs.Gifs          `bson:"gifs"`
}

func (model GifsByCategory) ToDto() GifsByCategoryDto {
	return GifsByCategoryDto{
		CategoryID: model.CategoryId.Hex(),
		Name:       model.Name,
		Gifs:       model.Gifs.ToDto(),
	}
}
