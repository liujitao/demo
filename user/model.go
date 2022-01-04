package user

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type UserModel struct {
    ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
    UserName  string             `json:"user_name" bson:"user_name"`
    RealName  string             `json:"real_name" bson:"real_name"`
    Mobile    string             `json:"mobile" bson:"mobile"`
    Email     string             `json:"email" bson:"email"`
    Password  string             `json:"password" bson:"password"`
    CreateAt  time.Time          `json:"create_at" bson:"create_at"`
    UpdateAt  time.Time          `json:"update_at" bson:"update_at"`
    CreateMan string             `kjson:"create_man" bson:"create_man"`
    UpdateMan string             `json:"update_man" bson:"update_man"`
    //Team     primitive.ObjectID `bson:"team_id" json:"team_id"`
    //Role     []string           `bson:"roles" json:"roles"`
    //Admin bool              `bson:"admin" json:"admin"`
}
