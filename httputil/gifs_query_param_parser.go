package httputil

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"net/url"
	"strconv"
)

type GifsQueryParamsParser struct {
	baseQueryParamsParser
}

func NewGifsApiQueryParamParser() *GifsQueryParamsParser {
	return &GifsQueryParamsParser{
		baseQueryParamsParser{},
	}
}

func (p *GifsQueryParamsParser) LoadValues(values url.Values) {
	p.baseQueryParamsParser.LoadValues(values)
}

func (p GifsQueryParamsParser) HasFilter() bool {
	return p.baseQueryParamsParser.HasFilter()
}

func (p GifsQueryParamsParser) GetFilter() (interface{}, error) {
	return p.baseQueryParamsParser.GetFilter(handleGifsApiFieldFilter)
}

func handleGifsApiFieldFilter(fieldName, operator, value string, filter bson.M) error {
	switch fieldName {
	case "name":
		filter[fieldName] = bson.M{
			operator: value, "$options": "i",
		}
		return nil
	case "isFavourite":
		boolValue, errBool := strconv.ParseBool(value)
		if errBool != nil {
			return fmt.Errorf("invalid filter value type for field 'isFavourite'")
		}
		filter[fieldName] = bson.M{
			operator: boolValue,
		}
		return nil
	}

	filter[fieldName] = bson.M{
		operator: value,
	}
	return nil
}
