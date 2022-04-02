package team

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type TeamModel struct {
    _ID       primitive.ObjectID `json:"_id" bson:"_id"`
    ID        string             `json:"id" bson:"id"`
    ParentID  string             `json:"parent_id" bson:"parent_id"`
    TeamName  string             `json:"team_name" bson:"team_name"`
    Describe  string             `json:"describe" bson:"describe"`
    Active    int64              `json:"active" bson:"active"`
    CreateAt  time.Time          `json:"create_at" bson:"create_at"`
    UpdateAt  time.Time          `json:"update_at" bson:"update_at"`
    CreateMan string             `json:"create_man" bson:"create_man"`
    UpdateMan string             `json:"update_man" bson:"update_man"`
    User      []string           `json:"user_id" bson:"user_id"`
    UserName  []string           `json:"user_name" bson:"user_name,omitempty"`
}
