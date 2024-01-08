package tests_usinc_custom_mocks

import (
	"context"
	"errors"
	"gifmanager-backend/dal"
	"reflect"
)

type MockDal struct {
	err          error
	insertResult *dal.InsertResult
	deleteResult *dal.DeleteResult
	updateResult *dal.UpdateResult
	findResult   interface{}
}

func NewMockDal() *MockDal {
	return &MockDal{}
}

func (m *MockDal) WithMockedInsertResult(result *dal.InsertResult) *MockDal {
	m.insertResult = result
	return m
}

func (m *MockDal) WithMockedDeleteResult(result *dal.DeleteResult) *MockDal {
	m.deleteResult = result
	return m
}

func (m *MockDal) WithMockedFindResult(result any) *MockDal {
	m.findResult = result
	return m
}

func (m *MockDal) WithError(err error) *MockDal {
	m.err = err
	return m
}

func (m *MockDal) FindByID(ctx context.Context, collection string, id string, result any) error {
	if m.err != nil {
		return m.err
	}

	return setFindResult(result, m.findResult)
}

func (m *MockDal) Disconnect(ctx context.Context) error {
	return m.err
}

func (m *MockDal) Insert(ctx context.Context, collection string, document []any) (*dal.InsertResult, error) {
	return m.insertResult, m.err
}

func (m *MockDal) Find(ctx context.Context, collection string, findArguments dal.FindArguments, result any) error {
	if m.err != nil {
		return m.err
	}

	return setFindResult(result, m.findResult)
}

func (m *MockDal) Delete(ctx context.Context, collection string, filter any) (*dal.DeleteResult, error) {
	return m.deleteResult, m.err
}

func (m *MockDal) FindAndDeleteByID(ctx context.Context, collection string, id string, result any) error {
	if m.err != nil {
		return m.err
	}

	return setFindResult(result, m.findResult)
}

func (m *MockDal) Aggregate(ctx context.Context, collection string, pipeline []any, result any) error {
	if m.err != nil {
		return m.err
	}

	return setFindResult(result, m.findResult)
}

func (m *MockDal) Update(ctx context.Context, collection string, filter any, update any, optionFuncs ...dal.UpdateOptionsFunc) (*dal.UpdateResult, error) {
	return m.updateResult, m.err
}

func (m *MockDal) UpdateByID(ctx context.Context, collection string, id string, update any, optionFuncs ...dal.UpdateOptionsFunc) (*dal.UpdateResult, error) {
	return m.updateResult, m.err
}

func setFindResult(obj any, findResult any) error {
	resultValue := reflect.ValueOf(obj)
	if resultValue.Kind() != reflect.Ptr || resultValue.IsNil() {
		return errors.New("result should be a non-nil pointer")
	}

	// Get the value that the pointer points to
	resultElem := resultValue.Elem()

	// Create a new value of the same type as the pointed value
	newValue := reflect.New(resultElem.Type()).Elem()
	newValue.Set(reflect.ValueOf(findResult))
	// Set the new value to the result using reflection
	resultElem.Set(newValue)
	return nil
}
