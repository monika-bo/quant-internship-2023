package mock_tests_using_library

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gifmanager-backend/dal"
	"gifmanager-backend/gifs"
	"gifmanager-backend/httputil"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	testifyhttp "github.com/stretchr/testify/http"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateGifHandler_ExpectedStatusCreated(t *testing.T) {
	// 1.ARRANGE
	userID := primitive.NewObjectID()
	categoryID := primitive.NewObjectID()

	expectedGif := gifs.Gif{
		Name:       t.Name(),
		URL:        "gif-url",
		CategoryId: categoryID,
		UserId:     userID,
		IsFavorite: false,
	}
	expectedInsertResult := &dal.InsertResult{
		InsertedDocumentsCount: 1,
	}

	// Mock DAL (Data Access Layer): Creates a mocked implementation of the DAL using testify/mock testing framework
	// and sets expectations on the methods that will be called. It specifies the expected arguments and return values for the Insert method
	// for more info on the framework check the official documentation https://github.com/stretchr/testify#mock-package
	mockedDal := dal.NewMockDAL(t)
	// expect that the Insert method will be called once with exactly the same arguments listed below
	mockedDal.On(
		"Insert",
		// we are not interested in the first argument that's why it can be anything
		mock.Anything,
		// the second argument must be 'gifs'
		dal.CollGifs,
		// as the third argument we pass a function that
		// validates that the actual third argument is what we expect
		mock.MatchedBy(
			func(actualGifs []any) bool {
				// in our case we validate that Insert method was called with array of gifs with only one element
				if len(actualGifs) != 1 {
					return false
				}
				actualGif := actualGifs[0].(gifs.Gif)

				// then we compare the actual properties with the expected ones to validate that they're the same
				return actualGif.Name == expectedGif.Name &&
					actualGif.UserId == userID &&
					actualGif.URL == expectedGif.URL &&
					actualGif.CategoryId == expectedGif.CategoryId &&
					actualGif.IsFavorite == expectedGif.IsFavorite &&
					actualGif.ID.IsZero() == false
			}),
		// as a final step we will have to specify the return values of this method call
		// in our case we return the expectedInsertResult and nil (nil represents the error)
	).Return(expectedInsertResult, nil)

	expectedUpdateResult := &dal.UpdateResult{
		MatchedCount: 1,
	}
	mockedDal.On(
		"UpdateByID",
		mock.Anything,
		dal.CollCategories,
		expectedGif.CategoryId.Hex(),
		bson.M{"$inc": bson.M{"gifCount": 1}},
	).Return(expectedUpdateResult, nil)

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

	// inject the mockedDal when we instantiate gif's api
	// remember that this mockedDal is DAL implementation that does nothing except it returns the predefined InsertResult when Insert method is called
	api := gifs.NewGifApi(mockedDal, httputil.NewGifsApiQueryParamParser())

	// 2.ACT
	api.CreateGifHandler(&responseRecorder, request)

	// 3.ASSERT
	assert.Equal(t, http.StatusCreated, responseRecorder.StatusCode)
}

func TestGetGifsHandler_ExpectedOk(t *testing.T) {
	// 1.ARRANGE
	userID := primitive.NewObjectID()
	categoryID := primitive.NewObjectID()

	expectedGif := gifs.Gif{
		ID:         primitive.NewObjectID(),
		Name:       "gif1",
		URL:        "gifUrl",
		IsFavorite: false,
		UserId:     userID,
		CategoryId: categoryID,
	}

	expectedFindArgs := dal.FindArguments{
		Filter: bson.M{"userId": expectedGif.UserId},
	}

	mockedDal := dal.NewMockDAL(t)
	mockedDal.On("Find",
		mock.Anything,
		dal.CollGifs,
		expectedFindArgs,
		mock.Anything,
	).
		Return(nil)

	request := httptest.NewRequest(http.MethodGet, "/gifs", nil)
	requestContext := context.WithValue(context.Background(), "userID", expectedGif.UserId)
	request = request.WithContext(requestContext)

	responseRecorder := httptest.NewRecorder()

	api := gifs.NewGifApi(mockedDal, httputil.NewGifsApiQueryParamParser())

	// 2.ACT
	api.GetGifsHandler(responseRecorder, request)

	// 3.ASSERT
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	var gifDTOs gifs.GifDtos
	body := responseRecorder.Body
	require.Nil(t, json.NewDecoder(body).Decode(&gifDTOs))
	assert.Len(t, gifDTOs, 0)
}

func TestGetGifsHandler_FindReturnsError_ExpectedInternalServerError(t *testing.T) {
	mockedDal := dal.NewMockDAL(t)
	mockedDal.On("Find", mock.Anything, dal.CollGifs, mock.Anything, mock.Anything).
		Return(fmt.Errorf("database is down"))

	request := httptest.NewRequest(http.MethodGet, "/gifs", nil)
	requestContext := context.WithValue(context.Background(), "userID", primitive.NewObjectID())
	request = request.WithContext(requestContext)

	responseRecorder := httptest.NewRecorder()

	api := gifs.NewGifApi(mockedDal, httputil.NewGifsApiQueryParamParser())
	api.GetGifsHandler(responseRecorder, request)

	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	assert.Equal(t, gifs.ErrFindingGifs, responseRecorder.Body.String())
}

func TestDeleteGifHandler_StatusNoContent(t *testing.T) {
	gifID := primitive.NewObjectID()
	mockedDal := dal.NewMockDAL(t)
	mockedDal.On("FindAndDeleteByID", mock.Anything, dal.CollGifs, gifID.Hex(), mock.Anything).
		Return(nil)

	request := httptest.NewRequest(http.MethodDelete, "/gifs", nil)
	requestContext := context.WithValue(context.Background(), "userID", primitive.NewObjectID())
	request = request.WithContext(requestContext)
	request = mux.SetURLVars(request, map[string]string{
		"id": gifID.Hex(),
	})

	responseRecorder := httptest.NewRecorder()

	api := gifs.NewGifApi(mockedDal, httputil.NewGifsApiQueryParamParser())
	api.DeleteGifHandler(responseRecorder, request)

	assert.Equal(t, http.StatusNoContent, responseRecorder.Code)
}
