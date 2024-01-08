package groups

import "go.mongodb.org/mongo-driver/bson/primitive"

type Group struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name   string             `bson:"name" json:"name"`
	UserId primitive.ObjectID `bson:"user_id" json:"user_id"`
	// store the Users IDs
	Contacts []string `bson:"contacts" json:"contacts"`
}
