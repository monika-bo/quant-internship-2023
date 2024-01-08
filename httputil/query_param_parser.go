package httputil

import (
	"net/url"
)

type QueryParamsParser interface {
	LoadValues(values url.Values)

	HasFilter() bool
	GetFilter() (interface{}, error)
}
