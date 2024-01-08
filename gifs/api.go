package gifs

import (
	"encoding/json"
	"fmt"
	"gifmanager-backend/dal"
	"gifmanager-backend/httputil"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type Api struct {
	Dal               dal.DAL
	QueryParamsParser httputil.QueryParamsParser
}

func NewGifApi(dal dal.DAL, parser httputil.QueryParamsParser) *Api {
	return &Api{
		Dal:               dal,
		QueryParamsParser: parser,
	}
}

func (api Api) InitializeEndpoints(route *mux.Router) {
	route.
		Path("/gifs").
		Methods(http.MethodPost).
		Handler(http.HandlerFunc(api.CreateGifHandler))
	route.
		Path("/gifs/{id}").
		Methods(http.MethodPut).
		Handler(http.HandlerFunc(api.UpdateGifHandler))
	route.
		Path("/gifs/{id}").
		Methods(http.MethodDelete).
		Handler(http.HandlerFunc(api.DeleteGifHandler))

	route.
		Path("/gifs").
		Methods(http.MethodGet).
		Handler(http.HandlerFunc(api.GetGifsHandler))
}

func (api Api) CreateGifHandler(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	userID := ctx.Value("userID")

	var gifRequest GifRequest

	if decodeErr := json.NewDecoder(request.Body).
		Decode(&gifRequest); decodeErr != nil {
		httputil.WriteHttpError(writer, http.StatusBadRequest, fmt.Sprintf(ErrDecodingGifFmt, decodeErr.Error()))
		return
	}

	gif := gifRequest.ToModel()
	gif.ID = primitive.NewObjectID()
	gif.UserId = userID.(primitive.ObjectID)
	if _, errInsert := api.Dal.Insert(ctx, dal.CollGifs, []any{gif}); errInsert != nil {
		fmt.Println(errInsert.Error())
		httputil.WriteHttpError(writer, http.StatusInternalServerError, ErrInsertingGifs)
		return
	}

	update := bson.M{"$inc": bson.M{"gifCount": 1}}
	if _, errUpdating := api.Dal.UpdateByID(ctx, dal.CollCategories, gif.CategoryId.Hex(), update); errUpdating != nil {
		fmt.Println(errUpdating.Error())
		httputil.WriteHttpError(writer, http.StatusInternalServerError, ErrUpdatingCategoriesCount)
		return
	}

	dto := gif.ToDto()
	if errEncode := json.NewEncoder(writer).Encode(&dto); errEncode != nil {
		fmt.Println(errEncode.Error())
		httputil.WriteHttpError(writer, http.StatusInternalServerError, ErrEncodingGifs)
		return
	}

	writer.WriteHeader(http.StatusCreated)
}

func (api Api) GetGifsHandler(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	userID := ctx.Value("userID").(primitive.ObjectID)

	api.QueryParamsParser.LoadValues(request.URL.Query())

	filter := make(bson.M)
	if api.QueryParamsParser.HasFilter() {
		queryFilter, err := api.QueryParamsParser.GetFilter()
		if err != nil {
			httputil.WriteHttpError(writer, http.StatusBadRequest, err.Error())
			return
		}
		filter = queryFilter.(bson.M)
	}
	filter["userId"] = userID

	findArgs := dal.NewFindArguments().
		WithFilter(filter)

	gifs := make(Gifs, 0)
	if err := api.Dal.Find(ctx, dal.CollGifs, *findArgs, &gifs); err != nil {
		fmt.Println(err)
		httputil.WriteHttpError(writer, http.StatusInternalServerError, ErrFindingGifs)
		return
	}

	if err := json.NewEncoder(writer).Encode(&gifs); err != nil {
		fmt.Println(err)
		httputil.WriteHttpError(writer, http.StatusInternalServerError, ErrEncodingGifs)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (api Api) DeleteGifHandler(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	id := mux.Vars(request)["id"]

	gifID, errObjId := primitive.ObjectIDFromHex(id)
	if errObjId != nil {
		httputil.WriteHttpError(writer, http.StatusBadRequest, fmt.Sprintf(ErrInvalidIDFmt, id))
		return
	}

	var deletedGif Gif
	if err := api.Dal.FindAndDeleteByID(ctx, dal.CollGifs, gifID.Hex(), &deletedGif); err != nil {
		httputil.WriteHttpError(writer, http.StatusInternalServerError, ErrDeletingGif)
		return
	}

	if deletedGif.CategoryId.IsZero() == false {
		update := bson.M{"$inc": bson.M{"gifCount": -1}}
		if _, errUpdating := api.Dal.UpdateByID(ctx, dal.CollCategories, deletedGif.CategoryId.Hex(), update); errUpdating != nil {
			httputil.WriteHttpError(writer, http.StatusInternalServerError, ErrUpdatingCategoriesCount)
			return
		}
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (api Api) UpdateGifHandler(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	userID := ctx.Value("userID").(primitive.ObjectID)
	id := mux.Vars(request)["id"]

	gifID, errObjId := primitive.ObjectIDFromHex(id)
	if errObjId != nil {
		httputil.WriteHttpError(writer, http.StatusBadRequest, fmt.Sprintf(ErrInvalidIDFmt, errObjId.Error()))
		return
	}

	var gifRequest GifRequest
	if decodeErr := json.NewDecoder(request.Body).
		Decode(&gifRequest); decodeErr != nil {
		httputil.WriteHttpError(writer, http.StatusBadRequest, fmt.Sprintf(ErrDecodingGifFmt, decodeErr.Error()))
		return
	}

	gif := gifRequest.ToModel()
	gif.UserId = userID
	update := bson.M{"$set": gif}

	filter := bson.M{"_id": gifID}
	result, errUpdating := api.Dal.Update(ctx, "gifs", filter, update)
	if errUpdating != nil {
		httputil.WriteHttpError(writer, http.StatusInternalServerError, ErrUpdatingGif)
		return
	}

	if result.MatchedCount == 0 {
		httputil.WriteHttpError(writer, http.StatusNotFound, fmt.Sprintf(ErrGifNotFoundFmt, id))
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
