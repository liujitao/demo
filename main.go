package main

import (
    "context"
    "demo/common"
    "demo/data"
    "demo/user"
    "encoding/json"
    "io/ioutil"
    "log"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"
)

// 定义变量
var Conf *common.Configure
var userHandler *user.UserHandler

func init() {
    // 读取配置文件
    content, err := ioutil.ReadFile("common/config.json")
    if err != nil {
        log.Fatal("Error when opening file: ", err)
    }

    err = json.Unmarshal(content, &Conf)
    if err != nil {
        log.Fatal("Error during Unmarshal(): ", err)
    }

    // 连接mongo
    mongo_uri := "mongodb://" + Conf.Mongo_user + ":" + Conf.Mongo_password + "@" + Conf.Mongo_host + ":" + Conf.Mongo_port
    ctx := context.Background()
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongo_uri))
    if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
        log.Fatal("Error connect to mongodb: ", err)
    }
    log.Println("Connected to Mongo")

    // 连接redis
    redisClient := redis.NewClient(&redis.Options{
        Addr:     Conf.Redis_host + ":" + Conf.Redis_port,
        Password: Conf.Redis_password,
        DB:       0,
    })

    status := redisClient.Ping(ctx)
    log.Println(status)

    db := client.Database(Conf.Mongo_database)

    // 初始化handler
    userHandler = user.MewUserHandler(ctx, db.Collection("user"), redisClient)

    // 导入数据
    if count, _ := db.Collection("user").CountDocuments(ctx, bson.M{}); count == 0 {
        _ = data.ImportUserData(ctx, db.Collection("user"))
    }
}

func main() {
    // gin工作模式
    if Conf.App_release_mode {
        gin.SetMode(gin.ReleaseMode)
    }

    router := gin.New()
    router.SetTrustedProxies([]string{})

    // 使用认证中间件
    authorized := router.Group("/")
    authorized.Use(userHandler.AuthMiddleWare())
    {
        // 用户
        authorized.POST("/user", userHandler.CreateUserHandler)
        authorized.GET("/user", userHandler.RetriveUserHandler)
        authorized.PUT("/user", userHandler.UpdateUserHandler)
        authorized.DELETE("/user", userHandler.DeleteUserHandler)
        authorized.POST("/user/change_password", userHandler.UserChanegePasswordHandler)
        authorized.GET("/user/logout", userHandler.UserLogoutHandler)
        authorized.GET("/user/blacklist", userHandler.UserBlackListHandler)
        authorized.POST("/user/blacklist", userHandler.UserBlackListAddHandler)
        authorized.DELETE("/user/blacklist", userHandler.UserBlackListRemoveHandler)
    }

    // 不使用认证中间件
    {
        router.POST("/user/register", userHandler.CreateUserHandler)
        router.POST("/user/login", userHandler.UserLoginHandler)
        router.GET("/user/refresh", userHandler.UserRefreshHandler)
    }

    router.Run(Conf.App_host + ":" + Conf.App_port)
}
