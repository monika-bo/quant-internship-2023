package benchmark_tests

import (
	"context"
	"gifmanager-backend/gifs"
	"gifmanager-backend/httputil"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/http/httptest"
	"testing"
)

// BenchmarkGetGifsHandler function: This is the benchmark function, prefixed with Benchmark to denote that it's a benchmark test.
// It accepts a *testing.B parameter.
func BenchmarkGetGifsHandler(b *testing.B) {
	// Create a mock request for the handler
	request, err := http.NewRequest(http.MethodGet, "/gifs", nil)
	requestContext := context.WithValue(context.Background(), "userID", primitive.NewObjectID())
	request = request.WithContext(requestContext)
	require.Nil(b, err)

	// Create a writer for the handler
	responseWriter := httptest.NewRecorder()

	api := gifs.NewGifApi(mongoDal, httputil.NewGifsApiQueryParamParser())
	// Running the handler in a loop: The loop for i := 0; i < b.N; i++ executes the handler b.N times.
	// The b.N value increases automatically by the testing framework to get a stable benchmark measurement.
	for i := 0; i < b.N; i++ {
		api.GetGifsHandler(responseWriter, request)
	}
}
