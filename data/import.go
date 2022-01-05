package data

import (
	"context"
	"demo/common"
	"demo/user"
	"encoding/json"
	"io/ioutil"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
导入用户数据
*/
func ImportUserData(ctx context.Context, collection *mongo.Collection) error {
	// json array to []struct
	content, err := ioutil.ReadFile("data/userData.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	var users []user.UserModel
	err = json.Unmarshal(content, &users)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	// []struct to []interface
	var document interface{}
	var documents []interface{}
	for _, u := range users {
		document = bson.D{
			{"_id", u.ID},
			{"user_name", u.UserName},
			{"real_name", u.RealName},
			{"mobile", u.Mobile},
			{"email", u.Email},
			{"password", common.SetPassword(u.Password)},
			{"create_at", u.CreateAt},
			{"update_at", u.CreateAt},
		}
		documents = append(documents, document)
	}

	// insert many
	_, err = collection.InsertMany(ctx, documents)
	if err != nil {
		return nil
	}
	return err
}
