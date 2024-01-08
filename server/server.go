package server

import (
	"context"
	"fmt"
	"gifmanager-backend/dal"
	"gifmanager-backend/users"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

type HttpServer interface {
	Run(address string) error
}

type Server struct {
	Handler http.Handler
}

type Api interface {
	InitializeEndpoints(route *mux.Router)
}

func NewServer(mongoDal dal.DAL, apis ...Api) Server {
	router := mux.NewRouter()
	router.Use(corsMiddleware())
	router.Methods(http.MethodOptions).
		HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {})

	loginRouter := router.PathPrefix("/login").Subrouter()
	loginRouter.Methods(http.MethodPost).Handler(http.HandlerFunc(users.NewLoginApi(mongoDal).LoginHandler))

	mainRouter := router.PathPrefix("").Subrouter()
	mainRouter.Use(authorizationMiddleware(mongoDal))

	for _, api := range apis {
		api.InitializeEndpoints(mainRouter)
	}

	return Server{
		Handler: router,
	}
}

func (m Server) Run(address string) error {
	return http.ListenAndServe(address, m.Handler)
}

func authorizationMiddleware(mongoDal dal.DAL) mux.MiddlewareFunc {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

			userName, password, ok := request.BasicAuth()
			if !ok {
				writer.WriteHeader(http.StatusUnauthorized)
				_, errWriter := writer.Write([]byte("missing authentication"))
				if errWriter != nil {
					fmt.Println(errWriter.Error())
				}
				return
			}
			result := make([]users.User, 0)
			ctx := request.Context()
			findArguments := dal.FindArguments{Filter: bson.M{"username": userName}}
			if err := mongoDal.Find(ctx, "users", findArguments, &result); err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				_, errWriter := writer.Write([]byte("unexpected error while trying to fetch the user"))
				if errWriter != nil {
					fmt.Println(errWriter.Error())
				}
				return
			}
			if len(result) == 0 {
				writer.WriteHeader(http.StatusUnauthorized)
				_, errWriter := writer.Write([]byte("user not found"))
				if errWriter != nil {
					fmt.Println(errWriter.Error())
				}
				return
			}

			if result[0].Password != password {
				writer.WriteHeader(http.StatusUnauthorized)
				_, errWriter := writer.Write([]byte("the password differs"))
				if errWriter != nil {
					fmt.Println(errWriter.Error())
				}
				return
			}

			request = request.WithContext(context.WithValue(ctx, "userID", result[0].ID))
			next.ServeHTTP(writer, request)
		})
	}
}

func corsMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Access-Control-Allow-Origin", "*")
			writer.Header().Set("Access-Control-Allow-Headers", "*")
			writer.Header().Set("Access-Control-Allow-Methods", "*")
			next.ServeHTTP(writer, request)
		})
	}
}
