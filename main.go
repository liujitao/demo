package main

import (
    "context"
    "demo/user"
    "encoding/json"
    "io/ioutil"
    "log"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"
)

type Configure struct {
    Mongo_host       string
    Mongo_port       string
    Mongo_user       string
    Mongo_password   string
    Mongo_database   string
    Redis_host       string
    Redis_port       string
    Redis_password   string
    App_host         string
    App_port         string
    App_release_mode bool
}

// variable
var Config Configure
var userHandler *user.UserHandler

func init() {
    // config
    content, err := ioutil.ReadFile("config.json")
    if err != nil {
        log.Fatal("Error when opening file: ", err)
    }

    err = json.Unmarshal(content, &Config)
    if err != nil {
        log.Fatal("Error during Unmarshal(): ", err)
    }

    // mongo
    mongo_uri := "mongodb://" + Config.Mongo_user + ":" + Config.Mongo_password + "@" + Config.Mongo_host + ":" + Config.Mongo_port
    ctx := context.Background()
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongo_uri))
    if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
        log.Fatal("Error connect to mongodb: ", err)
    }
    log.Println("Connected to Mongo")

    // redis
    redisClient := redis.NewClient(&redis.Options{
        Addr:     Config.Redis_host + ":" + Config.Redis_port,
        Password: Config.Redis_password,
        DB:       0,
    })

    status := redisClient.Ping()
    log.Println(status)

    db := client.Database(Config.Mongo_database)

    // handler new()
    userHandler = user.MewUserHandler(ctx, db.Collection("user"), redisClient)

    // import
    _ = importData(ctx, db.Collection("user"), "userData.json")
}

// bulk import
func importData(ctx context.Context, collection *mongo.Collection, file string) error {
    // json array to []struct
    content, err := ioutil.ReadFile("userData.json")
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
            {"password", u.Password},
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

func main() {
    if Config.App_release_mode {
        gin.SetMode(gin.ReleaseMode)
    }

    router := gin.New()
    router.SetTrustedProxies([]string{})

    // user
    router.POST("/user", userHandler.CreateUserHandler)
    router.GET("/user", userHandler.RetrieveUserHandler)
    router.PUT("/user", userHandler.UpdateUserHandler)
    router.DELETE("/user", userHandler.DeleteUserHandler)

    router.Run(Config.App_host + ":" + Config.App_port)
}
