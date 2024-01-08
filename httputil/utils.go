package httputil

import (
	"fmt"
	"net/http"
)

func WriteHttpError(writer http.ResponseWriter, statusCode int, errorMessage string) {
	writer.WriteHeader(statusCode)
	_, errWrite := writer.Write([]byte(errorMessage))
	if errWrite != nil {
		fmt.Println(errWrite.Error())
	}
}
