package integration_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"gifmanager-backend/categories"
	"gifmanager-backend/dal"
	"gifmanager-backend/gifs"
	"gifmanager-backend/httputil"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	testifyhttp "github.com/stretchr/testify/http"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateGifHandler_ExpectedStatusCreated(t *testing.T) {
	// 1. ARRANGE
	userID := primitive.NewObjectID()
	categoryID := primitive.NewObjectID()

	// create one category in the database with the same ID as the one in the request body
	category := categories.Category{
		ID:       categoryID,
		Name:     t.Name(),
		GifCount: 0,
		UserId:   userID,
	}
	_, errInsertCategory := mongoDal.Insert(context.Background(), dal.CollCategories, []any{category})
	require.Nil(t, errInsertCategory)

	// after the test finishes we have to clean all the documents that were inserted in the database
	defer func() {
		mongoDal.Delete(context.Background(), dal.CollGifs, bson.M{"userId": userID})
		mongoDal.Delete(context.Background(), dal.CollCategories, bson.M{"userId": userID})
	}()

	// create the request body
	requestBody := gifs.GifRequest{
		Name:       t.Name(),
		URL:        "gif-url",
		CategoryId: categoryID,
	}

	// we'll have to marshal it to bytes and transform it to a type that implements the io.Reader interface
	// so we can pass it to NewRequest() function
	bts, _ := json.Marshal(requestBody)
	reader := bytes.NewReader(bts)

	// create the request
	request := httptest.NewRequest(http.MethodPost, "/gifs", reader)
	// set the userID property in request's context
	requestContext := context.WithValue(context.Background(), "userID", userID)
	request = request.WithContext(requestContext)

	// create mock object that implements ResponseWriter interface
	responseRecorder := testifyhttp.TestResponseWriter{}

	// create new gifs api passing the global mongoDal variable and new query param parser
	api := gifs.NewGifApi(mongoDal, httputil.NewGifsApiQueryParamParser())

	// 2. ACT
	// call the handler under test
	api.CreateGifHandler(&responseRecorder, request)

	// 3. ASSERT
	assert.Equal(t, http.StatusCreated, responseRecorder.StatusCode)

	// first assert that response's body contains the DTO we expect
	var gifDTO gifs.GifDto
	body := responseRecorder.Output
	require.Nil(t, json.Unmarshal([]byte(body), &gifDTO))
	require.NotZero(t, gifDTO.ID)
	assert.Equal(t, requestBody.Name, gifDTO.Name)
	assert.Equal(t, requestBody.URL, gifDTO.URL)
	assert.Equal(t, requestBody.CategoryId.Hex(), gifDTO.CategoryID)

	// assert that a gif with that ID actually exists in the database
	var dbGif gifs.Gif
	errFind := mongoDal.FindByID(context.Background(), dal.CollGifs, gifDTO.ID, &dbGif)

	require.Nil(t, errFind)
	assert.Equal(t, requestBody.Name, dbGif.Name)
	assert.Equal(t, requestBody.URL, dbGif.URL)
	assert.Equal(t, requestBody.CategoryId, dbGif.CategoryId)
	assert.Equal(t, userID, dbGif.UserId)
	assert.False(t, dbGif.IsFavorite)

	// assert that category's gifCount property was increased
	var dbCategory categories.Category
	errFind = mongoDal.FindByID(context.Background(), dal.CollCategories, categoryID.Hex(), &dbCategory)
	require.Nil(t, errFind)
	assert.Equal(t, 1, dbCategory.GifCount)
}

func TestGetGifsHandler_ExpectedOk(t *testing.T) {
	// 1. ARRANGE
	userID := primitive.NewObjectID()
	// create one gif in the database with the same userID as the one in request's context
	expectedGif := gifs.Gif{
		ID:         primitive.NewObjectID(),
		Name:       t.Name(),
		URL:        "gifUrl",
		IsFavorite: false,
		UserId:     userID,
		CategoryId: primitive.NewObjectID(),
	}

	_, errInsert := mongoDal.Insert(context.Background(), dal.CollGifs, []any{
		expectedGif,
	})
	require.Nil(t, errInsert)

	// after the test finishes we have to clean all the documents that were inserted in the database
	defer func() {
		mongoDal.Delete(context.Background(), dal.CollGifs, bson.M{"userId": userID})
	}()

	// create the request
	request := httptest.NewRequest(http.MethodGet, "/gifs", nil)
	// set the userID in request's context
	requestContext := context.WithValue(context.Background(), "userID", userID)
	request = request.WithContext(requestContext)

	// create mock object that implements ResponseWriter interface
	responseRecorder := httptest.NewRecorder()
	// create new gifs api passing the global mongoDal variable and new query param parser
	api := gifs.NewGifApi(mongoDal, httputil.NewGifsApiQueryParamParser())

	// 2. ACT
	// call the handler under test
	api.GetGifsHandler(responseRecorder, request)

	// 3. ASSERT
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	var gifDTOs gifs.GifDtos
	body := responseRecorder.Body

	require.Nil(t, json.NewDecoder(body).Decode(&gifDTOs))
	require.Len(t, gifDTOs, 1)
	assert.Equal(t, expectedGif.Name, gifDTOs[0].Name)
	assert.Equal(t, expectedGif.ID.Hex(), gifDTOs[0].ID)
	assert.Equal(t, expectedGif.URL, gifDTOs[0].URL)
	assert.Equal(t, expectedGif.CategoryId.Hex(), gifDTOs[0].CategoryID)
}

func TestGetGifsHandler_FilterByFavourite_ExpectedOk(t *testing.T) {
	// 1.ARRANGE
	userID := primitive.NewObjectID()
	categoryID := primitive.NewObjectID()

	// this Gif belongs to the user making the request and is marked as favourite
	favouriteGif := gifs.Gif{
		ID:         primitive.NewObjectID(),
		Name:       t.Name(),
		URL:        "gifUrl",
		IsFavorite: true,
		UserId:     userID,
		CategoryId: categoryID,
	}

	// this gif belongs to the user making the request and is not marked as favourite
	ordinaryGif := gifs.Gif{
		ID:         primitive.NewObjectID(),
		Name:       t.Name(),
		URL:        "gifUrl2",
		IsFavorite: false,
		UserId:     userID,
		CategoryId: categoryID,
	}

	// this gif doesn't belong to the user making the request but is marked as favourite
	favouriteGifDifferentUser := gifs.Gif{
		ID:         primitive.NewObjectID(),
		Name:       t.Name(),
		URL:        "gifUrl2",
		IsFavorite: false,
		UserId:     primitive.NewObjectID(),
		CategoryId: primitive.NewObjectID(),
	}
	// only the first gif should be returned because it both belongs to the user and is marked as favourite

	// after the test finishes we have to clean all the documents that were inserted in the database
	defer func() {
		mongoDal.Delete(context.Background(), dal.CollGifs, bson.M{"userId": userID})
		mongoDal.Delete(context.Background(), dal.CollCategories, bson.M{"userId": userID})
		mongoDal.Delete(context.Background(), dal.CollGifs, bson.M{"userId": favouriteGifDifferentUser.UserId})
	}()

	// insert all gifs in the database
	_, errInsert := mongoDal.Insert(context.Background(), dal.CollGifs, []any{
		favouriteGif, ordinaryGif, favouriteGifDifferentUser,
	})
	require.Nil(t, errInsert)

	// create one category in the database with the same ID as the one in user's gifs
	category := categories.Category{
		Name:   t.Name(),
		ID:     categoryID,
		UserId: userID,
	}
	_, errInsert = mongoDal.Insert(context.Background(), dal.CollCategories, []any{
		category,
	})
	require.Nil(t, errInsert)

	// create the request
	request := httptest.NewRequest(http.MethodGet, "/gifs?filter=isFavourite-$eq-true", nil)
	// set the userID in request's context
	requestContext := context.WithValue(context.Background(), "userID", favouriteGif.UserId)
	request = request.WithContext(requestContext)

	// create mock object that implements ResponseWriter interface
	responseRecorder := httptest.NewRecorder()

	// create new gifs api passing the global mongoDal variable and new query param parser
	api := gifs.NewGifApi(mongoDal, httputil.NewGifsApiQueryParamParser())

	// 2.ACT
	// call the handler under test
	api.GetGifsHandler(responseRecorder, request)

	// 3.ASSERT
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	var gifDTOs gifs.GifDtos
	body := responseRecorder.Body

	require.Nil(t, json.NewDecoder(body).Decode(&gifDTOs))
	require.Len(t, gifDTOs, 1)
	assert.Equal(t, favouriteGif.Name, gifDTOs[0].Name)
	assert.Equal(t, favouriteGif.ID.Hex(), gifDTOs[0].ID)
	assert.Equal(t, favouriteGif.URL, gifDTOs[0].URL)
	assert.Equal(t, favouriteGif.CategoryId.Hex(), gifDTOs[0].CategoryID)
}

func TestDeleteGifHandler_StatusNoContent(t *testing.T) {
	// 1.ARRANGE
	userID := primitive.NewObjectID()
	categoryID := primitive.NewObjectID()

	// after the test finishes we have to clean all the documents that were inserted in the database
	defer func() {
		mongoDal.Delete(context.Background(), dal.CollGifs, bson.M{"userId": userID})
		mongoDal.Delete(context.Background(), dal.CollCategories, bson.M{"userId": userID})
	}()

	expectedGif := gifs.Gif{
		ID:         primitive.NewObjectID(),
		Name:       t.Name(),
		URL:        "gifUrl",
		IsFavorite: false,
		UserId:     userID,
		CategoryId: categoryID,
	}

	_, errInsertGif := mongoDal.Insert(context.Background(), dal.CollGifs, []any{
		expectedGif,
	})
	require.Nil(t, errInsertGif)

	// create one category in the database with the same ID as the one in user's gif
	category := categories.Category{
		ID:       categoryID,
		Name:     t.Name(),
		GifCount: 1,
		UserId:   userID,
	}
	_, errInsertCategory := mongoDal.Insert(context.Background(), dal.CollCategories, []any{
		category,
	})
	require.Nil(t, errInsertCategory)

	// create the request
	request := httptest.NewRequest(http.MethodDelete, "/gifs", nil)
	// set the userID in request's context
	requestContext := context.WithValue(context.Background(), "userID", userID)
	request = request.WithContext(requestContext)
	// set the id path variable to specify which gif to be deleted
	request = mux.SetURLVars(request, map[string]string{
		"id": expectedGif.ID.Hex(),
	})

	// create mock object that implements ResponseWriter interface
	responseRecorder := httptest.NewRecorder()

	// create new gifs api passing the global mongoDal variable and new query param parser
	api := gifs.NewGifApi(mongoDal, httputil.NewGifsApiQueryParamParser())

	// 2.ACT
	// call the method under test
	api.DeleteGifHandler(responseRecorder, request)

	// 3.ASSERT
	assert.Equal(t, http.StatusNoContent, responseRecorder.Code)

	// assert that the gif in the database is no longer present
	var dbGif gifs.Gif
	errFind := mongoDal.FindByID(context.Background(), dal.CollGifs, expectedGif.ID.Hex(), &dbGif)

	require.NotNil(t, errFind)
	assert.Equal(t, mongo.ErrNoDocuments, errFind)

	// assert that the gifCount property of the category is decreased
	var dbCategory categories.Category
	errFind = mongoDal.FindByID(context.Background(), dal.CollCategories, categoryID.Hex(), &dbCategory)
	require.Nil(t, errFind)
	assert.Equal(t, 0, dbCategory.GifCount)
}
