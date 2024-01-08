package users

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserName string             `bson:"username"json:"username"`
	Password string             `bson:"password" json:"password"`
}
type UserDTO struct {
	ID       primitive.ObjectID `json:"id,omitempty"`
	UserName string             `json:"username"`
}

func (u User) ToDTO() UserDTO {
	return UserDTO{
		ID:       u.ID,
		UserName: u.UserName,
	}
}
