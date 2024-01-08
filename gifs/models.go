package gifs

import "go.mongodb.org/mongo-driver/bson/primitive"

type Gif struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Name       string             `bson:"name"`
	URL        string             `bson:"url"`
	IsFavorite bool               `bson:"isFavourite"`
	UserId     primitive.ObjectID `bson:"userId"`
	CategoryId primitive.ObjectID `bson:"categoryId"`
}

func (gif Gif) ToDto() GifDto {
	return GifDto{
		ID:         gif.ID.Hex(),
		Name:       gif.Name,
		URL:        gif.URL,
		CategoryID: gif.CategoryId.Hex(),
	}
}

type Gifs []Gif

func (gifs Gifs) ToDto() GifDtos {
	dtos := make(GifDtos, 0, len(gifs))
	for i, gif := range gifs {
		dtos[i] = gif.ToDto()
	}
	return dtos
}
