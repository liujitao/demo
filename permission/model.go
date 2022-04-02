package permission

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type MenuModel struct {
    _ID       primitive.ObjectID `json:"_id" bson:"_id"`
    ID        string             `json:"id" bson:"id"`
    ParentID  string             `json:"parent_id" bson:"parent_id"`
    OrderID   int64              `json:"order" bson:"order"`
    MenuName  string             `json:"menu_name" bson:"menu_name"`
    Type      int64              `json:"type" bson:"type"`
    Path      string             `json:"path" bson:"path"`
    Component string             `json:"component" bson:"component"`
    Redirect  string             `json:"redirect" bson:"redirect"`
    Title     string             `json:"title" bson:"title"`
    Icon      string             `json:"icon" bson:"icon"`
    Hidden    int64              `json:"hidden" bson:"hidden"`
    CreateAt  time.Time          `json:"create_at" bson:"create_at"`
    UpdateAt  time.Time          `json:"update_at" bson:"update_at"`
    CreateMan string             `json:"create_man" bson:"create_man"`
    UpdateMan string             `json:"update_man" bson:"update_man"`
}
