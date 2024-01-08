package gifs

const (
	ErrInsertingGifs           = "error encountered on inserting the gifs"
	ErrUpdatingCategoriesCount = "error encountered on updating the categories count"
	ErrEncodingGifs            = "error encountered on encoding gifs"
	ErrDecodingGifFmt          = "error while decoding the request: %s"
	ErrFindingGifs             = "error encountered while retrieving favorite gifs"

	ErrInvalidIDFmt   = "invalid id: %s"
	ErrDeletingGif    = "error encountered deleting the gif"
	ErrUpdatingGif    = "error encountered on updating the gif"
	ErrGifNotFoundFmt = "gif with id %s does not exist"
)
