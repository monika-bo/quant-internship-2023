package tests_usinc_custom_mocks

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
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateGifHandler_ExpectedStatusCreated(t *testing.T) {
	// 1.ARRANGE
	categoryID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// we are going to mock the execution of dal's Insert method by using our custom mocked implementation of the DAL interface
	// first we have to create the mocked result that we want our Insert method to return
	mockedInsertResult := &dal.InsertResult{
		InsertedDocumentsCount: 1,
	}
	// then create a new instance of our custom MockDal and set the InsertResult property
	mockedDal := NewMockDal().
		WithMockedInsertResult(mockedInsertResult)

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
	// first assert that response's body contains the DTO we expect
	var gifDTO gifs.GifDto
	body := responseRecorder.Output
	require.Nil(t, json.Unmarshal([]byte(body), &gifDTO))
	require.NotZero(t, gifDTO.ID)
	assert.Equal(t, requestBody.Name, gifDTO.Name)
	assert.Equal(t, requestBody.URL, gifDTO.URL)
	assert.Equal(t, requestBody.CategoryId.Hex(), gifDTO.CategoryID)
}

func TestGetGifsHandler_ExpectedOk(t *testing.T) {
	// 1.ARRANGE
	categoryID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	expectedGif := gifs.Gif{
		ID:         primitive.NewObjectID(),
		Name:       t.Name(),
		URL:        "gifUrl",
		IsFavorite: false,
		UserId:     userID,
		CategoryId: categoryID,
	}

	// we are going to mock the execution of dal's Find method by using our custom mocked implementation of the DAL interface
	// first we have to create the mocked result that we want our Find method to return
	mockedFindResult := gifs.Gifs{expectedGif}
	// then create a new instance of our custom MockDal and set the FindResult property
	mockedDal := NewMockDal().
		WithMockedFindResult(mockedFindResult)

	// create the request
	request := httptest.NewRequest(http.MethodGet, "/gifs", nil)
	// set the userID in request's context
	requestContext := context.WithValue(context.Background(), "userID", userID)
	request = request.WithContext(requestContext)

	// create mock object that implements ResponseWriter interface
	responseRecorder := httptest.NewRecorder()
	// create new gifs api passing the global mongoDal variable and new query param parser
	api := gifs.NewGifApi(mockedDal, httputil.NewGifsApiQueryParamParser())

	// 2.ACT
	// call the handler under test
	api.GetGifsHandler(responseRecorder, request)

	// 3.ASSERT
	require.Equal(t, http.StatusOK, responseRecorder.Code)

	var gifDTOs gifs.GifDtos
	body := responseRecorder.Body

	require.Nil(t, json.NewDecoder(body).Decode(&gifDTOs))
	require.Len(t, gifDTOs, len(mockedFindResult))
	assert.Equal(t, expectedGif.Name, gifDTOs[0].Name)
	assert.Equal(t, expectedGif.ID.Hex(), gifDTOs[0].ID)
	assert.Equal(t, expectedGif.URL, gifDTOs[0].URL)
	assert.Equal(t, expectedGif.CategoryId.Hex(), gifDTOs[0].CategoryID)
}

func TestGetGifsHandler_FindReturnsError_ExpectedInternalServerError(t *testing.T) {
	// 1.ARRANGE
	userID := primitive.NewObjectID()

	// we are going to mock the execution of dal's Find method by using our custom mocked implementation of the DAL interface
	// we want our Find method to return an error that's why we are initializing new MockDal and setting
	// the error property
	mockedDal := NewMockDal().
		WithError(fmt.Errorf("database is down"))

	// create the request
	request := httptest.NewRequest(http.MethodGet, "/gifs", nil)
	// set the userID in request's context
	requestContext := context.WithValue(context.Background(), "userID", userID)
	request = request.WithContext(requestContext)

	// create mock object that implements ResponseWriter interface
	responseRecorder := httptest.NewRecorder()
	// create new gifs api passing the global mongoDal variable and new query param parser
	api := gifs.NewGifApi(mockedDal, httputil.NewGifsApiQueryParamParser())

	// 2. ACT
	// call the handler under test
	api.GetGifsHandler(responseRecorder, request)

	// 3. ASSERT
	require.Equal(t, http.StatusInternalServerError, responseRecorder.Code)
	assert.Equal(t, gifs.ErrFindingGifs, responseRecorder.Body.String())
}

func TestDeleteGifHandler_StatusNoContent(t *testing.T) {
	// 1.ARRANGE
	userID := primitive.NewObjectID()
	expectedGif := gifs.Gif{
		ID:         primitive.NewObjectID(),
		Name:       t.Name(),
		URL:        "gifUrl",
		IsFavorite: false,
		UserId:     userID,
		CategoryId: primitive.NewObjectID(),
	}
	mockedDal := NewMockDal().
		WithMockedFindResult(expectedGif)

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
	api := gifs.NewGifApi(mockedDal, httputil.NewGifsApiQueryParamParser())

	// 2.ACT
	// call the method under test
	api.DeleteGifHandler(responseRecorder, request)

	// 3.ASSERT
	assert.Equal(t, http.StatusNoContent, responseRecorder.Code)
}
