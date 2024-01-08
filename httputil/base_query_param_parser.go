package httputil

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"net/url"
	"strings"
)

type handleFieldFilterFunc func(fieldName, operator, value string, filter bson.M) error

type baseQueryParamsParser struct {
	values url.Values
}

func (b *baseQueryParamsParser) LoadValues(values url.Values) {
	b.values = values
}

func (b baseQueryParamsParser) HasFilter() bool {
	skip := b.values.Get("filter")
	return len(skip) > 0
}

func (b baseQueryParamsParser) GetFilter(handleFieldFilter handleFieldFilterFunc) (interface{}, error) {
	filterParam := b.values.Get("filter")
	filters := strings.Split(filterParam, ";")

	bsonFilter := make(bson.M, 0)
	for _, filter := range filters {
		filterParts := strings.Split(filter, "-")
		if len(filterParts) != 3 {
			return nil, fmt.Errorf("invalid filter format: %s", filter)
		}
		fieldName := filterParts[0]
		operator := filterParts[1]
		value := filterParts[2]

		if err := handleFieldFilter(fieldName, operator, value, bsonFilter); err != nil {
			return nil, err
		}
	}
	return bsonFilter, nil
}
