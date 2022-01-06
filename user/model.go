package user

import (
    "time"

    "github.com/dgrijalva/jwt-go"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type UserModel struct {
    ID        primitive.ObjectID `json:"_id" bson:"_id"`
    UserName  string             `json:"user_name" bson:"user_name"`
    RealName  string             `json:"real_name" bson:"real_name"`
    Mobile    string             `json:"mobile" bson:"mobile"`
    Email     string             `json:"email" bson:"email"`
    Password  string             `json:"password" bson:"password"`
    CreateAt  time.Time          `json:"create_at" bson:"create_at"`
    UpdateAt  time.Time          `json:"update_at" bson:"update_at"`
    CreateMan string             `json:"create_man,omitempty" bson:"create_man,omitempty"`
    UpdateMan string             `json:"update_man,omitempty" bson:"update_man,omitempty"`
    //Team     primitive.ObjectID `bson:"team_id" json:"team_id"`
    //Role     []string           `bson:"roles" json:"roles"`
    //Admin bool              `bson:"admin" json:"admin"`
}

type UserPasswordModel struct {
    ID          primitive.ObjectID `json:"_id"`
    OldPassword string             `json:"old_password"`
    NewPassword string             `json:"new_password"`
}

type UserLoginModel struct {
    UserName string `json:"user_name"`
    Password string `json:"password"`
}

// jwt
type CustomClaims struct {
    UserName string `json:"username"`
    jwt.StandardClaims
}

const JWT_SECRET = "7ffc6cc0-6dff-11ec-b5e4-97162d942513"
