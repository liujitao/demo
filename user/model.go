package user

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type UserModel struct {
    ID       primitive.ObjectID `bson:"_id" json:"_id"`
    UserName string             `bson:"user_name" json:"user_name"`
    RealName string             `bson:"real_name" json:"real_name"`
    Mobile   string             `bson:"mobile" json:"mobile"`
    Email    string             `bson:"email" json:"email"`
    Password string             `bson:"password" json:"password"`
    CreateAt time.Time          `bson:"create_at" json:"create_at"`
    UpdateAt time.Time          `bson:"update_at" json:"update_at"`
    //CreateMan string              `bson:"create_man" json:"create_man"`
    //UpdateMan string              `bson:"update_man" json:"update_man"`
    //Team     primitive.ObjectID `bson:"team_id" json:"team_id"`
    //Role     []string           `bson:"roles" json:"roles"`
    //Admin bool              `bson:"admin" json:"admin"`
}
