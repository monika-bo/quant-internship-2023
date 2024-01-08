package users

import (
	"context"
	"encoding/json"
	"fmt"
	"gifmanager-backend/dal"
	"gifmanager-backend/httputil"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"regexp"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type Api struct {
	Dal dal.DAL
}

func NewLoginApi(dal dal.DAL) *Api {
	return &Api{
		Dal: dal,
	}
}

func (api Api) InitializeEndpoints(route *mux.Router) {
	route.
		Path("/login").
		Methods(http.MethodPost).
		Handler(http.HandlerFunc(api.LoginHandler))

}

// Handle decoding of the login request
// Validate email and password
// Authenticate user
// Omit the password in the response
// Encode the user as JSON and write to the response

func (api Api) LoginHandler(writer http.ResponseWriter, request *http.Request) {

	var loginRequest LoginRequest
	if decodeErr := json.NewDecoder(request.Body).Decode(&loginRequest); decodeErr != nil {
		httputil.WriteHttpError(writer, http.StatusBadRequest, fmt.Sprintf("error while decoding the request: %s", decodeErr.Error()))
		return
	}
	if !isValidEmail(loginRequest.UserName) {
		httputil.WriteHttpError(writer, http.StatusBadRequest, fmt.Sprintf("invalid email address"))
		return
	}

	if !isValidPassword(loginRequest.Password) {
		httputil.WriteHttpError(writer, http.StatusBadRequest, fmt.Sprintf("invalid password. it should be at least 8 characters long and contain both a letter and a digit."))
		return
	}

	user, err := api.authenticateUser(loginRequest.UserName, loginRequest.Password)
	if err != nil {
		httputil.WriteHttpError(writer, http.StatusUnauthorized, fmt.Sprintf("authentication failed"))
		return
	}
	userDTO := user.ToDTO()

	if errEncode := json.NewEncoder(writer).Encode(userDTO); errEncode != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, errWrite := writer.Write([]byte(fmt.Sprintf("error encoding user: %s", errEncode.Error())))
		if errWrite != nil {
			fmt.Println(errWrite.Error())
		}
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (api Api) authenticateUser(userName, password string) (*User, error) {
	var result []User

	ctx := context.Background()

	findArguments := dal.FindArguments{Filter: bson.M{"username": userName}}
	if err := api.Dal.Find(ctx, dal.CollUsers, findArguments, &result); err != nil {
		return nil, err
	}
	// If either of these conditions is true, a new user is created. The password for the new user is hashed using bcrypt,
	//a new unique ObjectID is generated for the user, and the user is inserted into the database. The newly created user is then returned.

	if len(result) == 0 {
		newUser := User{
			ID:       primitive.NewObjectID(),
			UserName: userName,
			Password: password,
		}

		if _, err := api.Dal.Insert(ctx, dal.CollUsers, []interface{}{&newUser}); err != nil {
			return nil, err
		}
		return &newUser, nil
	}

	return &result[0], nil
}

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}
func isValidPassword(password string) bool {
	return len(password) >= 8 && containsLetter(password) && containsDigit(password)
}

func containsLetter(s string) bool {
	for _, char := range s {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			return true
		}
	}
	return false
}

func containsDigit(s string) bool {
	for _, char := range s {
		if char >= '0' && char <= '9' {
			return true
		}
	}
	return false
}
