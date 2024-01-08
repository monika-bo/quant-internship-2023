package integration_tests

import (
	"context"
	"gifmanager-backend/dal"
	"os"
	"testing"
)

var mongoDal dal.DAL

// In Go, TestMain is a special function that can be included in test files.
// It allows for custom setup and teardown logic that should run once before and after running any test functions in the same package.
func TestMain(m *testing.M) {
	// the setUp function is called before running the tests
	tearDown := setUp()
	// m.Run() runs the tests
	code := m.Run()
	// tearDown is a function that we are calling after the tests finish
	// its job is to release resources that are no longer in use
	// in our case it will just disconnect from the database
	tearDown()

	os.Exit(code)
}

func setUp() func() {

	// connect to the mongo but not to the gif-manager database
	// connect to a new database which we are going to use for the tests
	var err error
	mongoDal, err = dal.NewMongoDal(context.Background(), "mongodb://localhost:27017", dal.TestDbName)
	if err != nil {
		panic(err)
	}

	// return a tearDown() func
	return func() {
		// disconnect from the database
		if err := mongoDal.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}
}
