package user

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type UserModel struct {
    ID        primitive.ObjectID `json:"_id" bson:"_id"`
    UUID      string             `json:"uuid" bson:"uuid"`
    UserName  string             `json:"user_name" bson:"user_name"`
    RealName  string             `json:"real_name" bson:"real_name"`
    Mobile    string             `json:"mobile" bson:"mobile"`
    Email     string             `json:"email" bson:"email"`
    Password  string             `json:"password" bson:"password"`
    CreateAt  time.Time          `json:"create_at" bson:"create_at"`
    UpdateAt  time.Time          `json:"update_at" bson:"update_at"`
    CreateMan string             `json:"create_man" bson:"create_man"`
    UpdateMan string             `json:"update_man" bson:"update_man"`
    TeamName  string             `json:"team_name" bson:"team_name,omitempty"`
    RoleName  string             `json:"role_name" bson:"role_name,omitempty"`
}

type UserPasswordModel struct {
    UUID        string `json:"uuid" bson:"uuid"`
    OldPassword string `json:"old_password"`
    NewPassword string `json:"new_password"`
}

type UserLoginModel struct {
    UserName string `json:"user_name"`
    Password string `json:"password"`
}
