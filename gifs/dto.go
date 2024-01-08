package gifs

import "go.mongodb.org/mongo-driver/bson/primitive"

type GifRequest struct {
	Name        string             ` json:"name"`
	URL         string             ` json:"url"`
	CategoryId  primitive.ObjectID ` json:"categoryId"`
	IsFavourite bool               `json:"isFavourite"`
}

func (g GifRequest) ToModel() Gif {
	return Gif{
		Name:       g.Name,
		URL:        g.URL,
		IsFavorite: g.IsFavourite,
		CategoryId: g.CategoryId,
	}
}

type GifDto struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	CategoryID string `json:"categoryId"`
}

type GifDtos []GifDto
