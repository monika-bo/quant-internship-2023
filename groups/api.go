package groups

import (
	"encoding/json"
	"fmt"
	"gifmanager-backend/dal"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type Api struct {
	Dal dal.DAL
}

func NewGroupApi(dal dal.DAL) *Api {
	return &Api{
		Dal: dal,
	}
}

func (api Api) InitializeEndpoints(route *mux.Router) {
	route.
		Path("/groups").
		Methods(http.MethodPost).
		Handler(http.HandlerFunc(api.CreateGroupHandler))
	route.
		Path("/groups/{id}").
		Methods(http.MethodDelete).
		Handler(http.HandlerFunc(api.DeleteGroupHandler))
	route.
		Path("/groups/{id}").
		Methods(http.MethodPut).
		Handler(http.HandlerFunc(api.UpdateGroupHandler))

	route.
		Path("/groups").
		Methods(http.MethodGet).
		Handler(http.HandlerFunc(api.GetGroupHandler))

}

func (api Api) CreateGroupHandler(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	userID := ctx.Value("userID")

	var groupRequest GroupRequest
	if decodeErr := json.NewDecoder(request.Body).
		Decode(&groupRequest); decodeErr != nil {

		writer.WriteHeader(http.StatusBadRequest)
		_, errWrite := writer.Write([]byte(fmt.Sprintf("error while decoding the request: %s", decodeErr.Error())))
		if errWrite != nil {
			fmt.Println(errWrite.Error())
		}
		return
	}

	group := groupRequest.ToModel()
	group.UserId = userID.(primitive.ObjectID)
	_, errInsert := api.Dal.Insert(ctx, "groups", []any{group})
	if errInsert != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, errWrite := writer.Write([]byte("error encountered on inserting the groups"))
		if errWrite != nil {
			fmt.Println(errWrite.Error())
		}
		return
	}

	if errEncode := json.NewEncoder(writer).Encode(group); errEncode != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, errWrite := writer.Write([]byte("error encountered on encoding the group"))
		if errWrite != nil {
			fmt.Println(errWrite.Error())
		}
		return
	}

	writer.WriteHeader(http.StatusCreated)
}

func (api Api) GetGroupHandler(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	groups := make([]Group, 0)
	findArgs := dal.FindArguments{}

	if err := api.Dal.Find(ctx, "groups", findArgs, &groups); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, errWrite := writer.Write([]byte("error encountered on retrieving the groups"))

		if errWrite != nil {
			fmt.Println(errWrite.Error())
		}
		return
	}

	if errEncode := json.NewEncoder(writer).Encode(groups); errEncode != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, errWrite := writer.Write([]byte("error encountered on encoding the groups"))
		if errWrite != nil {
			fmt.Println(errWrite.Error())
		}
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (api Api) DeleteGroupHandler(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	params := mux.Vars(request)
	groupID, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		_, errWrite := writer.Write([]byte("invalid group ID"))
		if errWrite != nil {
			fmt.Println(errWrite.Error())
		}
		return
	}

	_, err = api.Dal.Delete(ctx, "groups", bson.M{"_id": groupID})
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, errWrite := writer.Write([]byte("error encountered while deleting the group"))
		if errWrite != nil {
			fmt.Println(errWrite.Error())
		}
		return
	}

	writer.WriteHeader(http.StatusNoContent)
	_, _ = writer.Write([]byte("group deleted successfully"))
}

func (api Api) UpdateGroupHandler(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	id := mux.Vars(request)["id"]

	ObjId, errObjId := primitive.ObjectIDFromHex(id)
	if errObjId != nil {

		writer.WriteHeader(http.StatusBadRequest)
		_, errWrite := writer.Write([]byte(fmt.Sprintf("invalid id specified: %s", id)))
		if errWrite != nil {
			fmt.Println(errWrite.Error())
		}
		return
	}

	var groupRequest GroupRequest
	if decodeErr := json.NewDecoder(request.Body).
		Decode(&groupRequest); decodeErr != nil {
		writer.WriteHeader(http.StatusBadRequest)
		_, errWrite := writer.Write([]byte(fmt.Sprintf("error while decoding the request: %s", decodeErr.Error())))
		if errWrite != nil {

			fmt.Println(errWrite.Error())
		}
		return
	}

	group := groupRequest.ToModel()

	group.UserId = ObjId

	update := bson.M{"$set": group}

	filter := bson.M{"_id": id}

	result, errUpdating := api.Dal.Update(ctx, "groups", filter, update)
	if errUpdating != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte("error encountered on updating the group"))
		return
	}

	if result.MatchedCount == 0 {
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte(fmt.Sprintf("group with id %s does not exist", id)))
		return
	}

	writer.WriteHeader(http.StatusOK)

}
