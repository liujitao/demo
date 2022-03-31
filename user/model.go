package user

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type UserModel struct {
    _ID       primitive.ObjectID `json:"_id" bson:"_id"`
    ID        string             `json:"id" bson:"id"`
    UserName  string             `json:"user_name" bson:"user_name"`
    Mobile    string             `json:"mobile" bson:"mobile"`
    Email     string             `json:"email" bson:"email"`
    Password  string             `json:"password" bson:"password"`
    Active    int64              `json:"active" bson:"active"`
    CreateAt  time.Time          `json:"create_at" bson:"create_at"`
    UpdateAt  time.Time          `json:"update_at" bson:"update_at"`
    CreateMan string             `json:"create_man" bson:"create_man"`
    UpdateMan string             `json:"update_man" bson:"update_man"`
    TeamName  string             `json:"team_name" bson:"team_name,omitempty"`
    RoleName  string             `json:"role_name" bson:"role_name,omitempty"`
}

type UserPasswordModel struct {
    ID          string `json:"id" bson:"id"`
    OldPassword string `json:"old_password"`
    NewPassword string `json:"new_password"`
}

type UserLoginModel struct {
    LoginID  string `json:"login_id"`
    Password string `json:"password"`
}
