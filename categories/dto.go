package categories

import "gifmanager-backend/gifs"

type CategoryRequest struct {
	Name string `json:"name"`
}

func (c CategoryRequest) ToModel() Category {
	return Category{
		Name: c.Name,
	}
}

type CategoryDto struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	GifsCount int    `json:"gifsCount"`
}

type GifsByCategoryDto struct {
	CategoryID string       `json:"categoryId,omitempty"`
	Name       string       `json:"name"`
	Gifs       gifs.GifDtos `json:"gifs"`
}
