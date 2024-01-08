package categories

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

func NewApi(dal dal.DAL, parser httputil.QueryParamsParser) *Api {
	return &Api{
		Dal:               dal,
		QueryParamsParser: parser,
	}
}

func (api Api) InitializeEndpoints(route *mux.Router) {
	route.
		Path("/categories").
		Methods(http.MethodPost).
		Handler(http.HandlerFunc(api.CreateCategoryHandler))
	route.
		Path("/categories/{id}").
		Methods(http.MethodPut).
		Handler(http.HandlerFunc(api.UpdateCategoryByIdHandler))
	route.
		Path("/categories/{id}").
		Methods(http.MethodDelete).
		Handler(http.HandlerFunc(api.DeleteCategoryByIdHandler))
	route.
		Path("/categories").
		Methods(http.MethodGet).
		Handler(http.HandlerFunc(api.GetCategoriesHandler))
	route.
		Path("/categories/gifs").
		Methods(http.MethodGet).
		Handler(http.HandlerFunc(api.GetGifsByCategory))
}

func (api Api) CreateCategoryHandler(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	userID := ctx.Value("userID")

	var categoryRequest CategoryRequest
	if decodeErr := json.NewDecoder(request.Body).
		Decode(&categoryRequest); decodeErr != nil {

		fmt.Println(decodeErr.Error())
		httputil.WriteHttpError(writer, http.StatusBadRequest, fmt.Sprintf("error while decoding the request: %s", decodeErr.Error()))
		return
	}

	category := categoryRequest.ToModel()
	category.ID = primitive.NewObjectID()
	category.UserId = userID.(primitive.ObjectID)

	if _, errInsert := api.Dal.Insert(ctx, dal.CollCategories, []any{category}); errInsert != nil {
		fmt.Println(errInsert.Error())
		httputil.WriteHttpError(writer, http.StatusInternalServerError, "error encountered on inserting the category")
		return
	}

	dto := category.ToDto()
	if errEncode := json.NewEncoder(writer).Encode(&dto); errEncode != nil {
		writer.WriteHeader(http.StatusInternalServerError)

		fmt.Println(errEncode.Error())
		httputil.WriteHttpError(writer, http.StatusInternalServerError, "error encountered on encoding the category")
		return
	}

	writer.WriteHeader(http.StatusCreated)
}

func (api Api) UpdateCategoryByIdHandler(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	id := mux.Vars(request)["id"]

	if _, errObjId := primitive.ObjectIDFromHex(id); errObjId != nil {
		httputil.WriteHttpError(writer, http.StatusBadRequest, fmt.Sprintf("invalid id specified: %s", id))
		return
	}

	var categoryRequest CategoryRequest
	if decodeErr := json.NewDecoder(request.Body).Decode(&categoryRequest); decodeErr != nil {
		httputil.WriteHttpError(writer, http.StatusInternalServerError, fmt.Sprintf("error while decoding the request: %s", decodeErr.Error()))
		return
	}

	category := categoryRequest.ToModel()

	update := bson.M{"$set": category}
	result, errUpdating := api.Dal.UpdateByID(ctx, dal.CollCategories, id, update)

	if errUpdating != nil {
		fmt.Println(errUpdating)
		httputil.WriteHttpError(writer, http.StatusInternalServerError, fmt.Sprintf("error while updating category"))
		return
	}

	if result.MatchedCount == 0 {
		httputil.WriteHttpError(writer, http.StatusNotFound, fmt.Sprintf("category with id %s does not exist", id))
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (api Api) GetCategoriesHandler(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	userID := ctx.Value("userID").(primitive.ObjectID)


	userIdFilter := bson.M{
		"userId": userID,
	}
	findArgs := dal.NewFindArguments().
		WithFilter(userIdFilter)

	var categories []Category
	if err := api.Dal.Find(ctx, dal.CollCategories, *findArgs, &categories); err != nil {
		fmt.Println(err)
		httputil.WriteHttpError(writer, http.StatusInternalServerError, "error encountered while retrieving categories")
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(categories); err != nil {
		fmt.Println(err)
		httputil.WriteHttpError(writer, http.StatusInternalServerError, "error encountered on encoding categories")
		return
	}

}

func (api Api) DeleteCategoryByIdHandler(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	id := mux.Vars(request)["id"]

	categoryID, errObjId := primitive.ObjectIDFromHex(id)
	if errObjId != nil {
		httputil.WriteHttpError(writer, http.StatusBadRequest, fmt.Sprintf("invalid id specified: %s", id))
		return
	}

	filter := bson.M{
		"_id": categoryID,
	}

	if _, err := api.Dal.Delete(ctx, dal.CollCategories, filter); err != nil {
		fmt.Println(err)
		httputil.WriteHttpError(writer, http.StatusInternalServerError, fmt.Sprintf("error encountered deleting the category"))
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (api Api) GetGifsByCategory(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	userID := ctx.Value("userID").(primitive.ObjectID)

	api.QueryParamsParser.LoadValues(request.URL.Query())

	pipeline := make([]any, 0)
	filter := []interface{}{bson.M{"userId": userID}}
	if api.QueryParamsParser.HasFilter() {
		queryFilter, err := api.QueryParamsParser.GetFilter()
		if err != nil {
			httputil.WriteHttpError(writer, http.StatusBadRequest, err.Error())
			return
		}

		filter = append(filter, queryFilter)
		pipeline = append(pipeline, bson.M{
			"$match": filter,
		})
	}

	var gifsByCategory []GifsByCategory
	pipeline = append(pipeline, getGifsByCategoriesPipeline(userID)...)

	if err := api.Dal.Aggregate(ctx, dal.CollGifs, pipeline, &gifsByCategory); err != nil {
		fmt.Println(err.Error())
		httputil.WriteHttpError(writer, http.StatusInternalServerError, "error encountered while retrieving gifs by category")
		return
	}

	if err := json.NewEncoder(writer).Encode(gifsByCategory); err != nil {
		fmt.Println(err.Error())
		httputil.WriteHttpError(writer, http.StatusInternalServerError, fmt.Sprintf("erron on encoding gifs"))
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func getGifsByCategoriesPipeline(userID primitive.ObjectID) []any {
	return []any{
		bson.M{"$match": bson.M{"user_id": userID}},
		bson.M{"$lookup": bson.M{
			"from":         "categories",
			"localField":   "category_id",
			"foreignField": "_id",
			"as":           "categories",
		}},
		bson.M{"$addFields": bson.M{
			"category": bson.M{"$arrayElemAt": bson.A{"$categories", 0}},
		}},
		bson.M{"$group": bson.M{"_id": "$category_id", "gifs": bson.M{"$push": "$$ROOT"}, "name": bson.M{"$first": "$category.name"}}},
		bson.M{"$project": bson.M{
			"category_id": "$_id",
			"gifs":        1,
			"name":        1,
		}},
	}
}
