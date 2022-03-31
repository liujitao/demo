package main

import (
    "context"
    "demo/common"
    "demo/role"
    "demo/team"
    "demo/user"
    "log"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"
)

// 定义变量
var userHandler *user.UserHandler
var teamHandler *team.TeamHandler
var roleHandler *role.RoleHandler

// 初始化
func init() {
    // 获取配置文件
    common.GetConfig()
    //log.Println(common.Conf)
    // log.Println(common.SetPassword("password"))

    // 连接mongo
    mongo_uri := "mongodb://" + common.Conf.Mongo_user + ":" + common.Conf.Mongo_password + "@" + common.Conf.Mongo_host + ":" + common.Conf.Mongo_port
    ctx := context.Background()
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongo_uri))
    if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
        log.Fatal("Error connect to mongodb: ", err)
    }
    log.Println("Connected to Mongo")

    // 连接redis
    redisClient := redis.NewClient(&redis.Options{
        Addr:     common.Conf.Redis_host + ":" + common.Conf.Redis_port,
        Password: common.Conf.Redis_password,
        DB:       0,
    })

    status := redisClient.Ping(ctx)
    log.Println(status)

    db := client.Database(common.Conf.Mongo_database)

    // 初始化handler
    userHandler = user.MewUserHandler(ctx, db.Collection("user"), redisClient)
    teamHandler = team.MewTeamHandler(ctx, db.Collection("team"))
    roleHandler = role.MewRoleHandler(ctx, db.Collection("role"))
}

func main() {
    // gin工作模式
    if common.Conf.App_release_mode {
        gin.SetMode(gin.ReleaseMode)
    }

    router := gin.New()
    router.SetTrustedProxies([]string{})
    unauthorized := router.Group("/v2")

    // 使用认证中间件
    authorized := router.Group("/v2")
    authorized.Use(userHandler.AuthMiddleWare())
    {
        // 用户
        authorized.POST("/user", userHandler.CreateUserHandler)
        authorized.GET("/user", userHandler.RetriveUserHandler)
        authorized.PUT("/user", userHandler.UpdateUserHandler)
        authorized.DELETE("/user", userHandler.DeleteUserHandler)
        authorized.GET("/user/list", userHandler.RetriveUserListHandler)
        authorized.POST("/user/change_password", userHandler.UserChanegePasswordHandler)
        authorized.GET("/user/logout", userHandler.UserLogoutHandler)
        authorized.GET("/user/refresh", userHandler.UserRefreshHandler)

        // 用户黑名单
        authorized.POST("/user/blacklist", userHandler.UserBlackListAddHandler)
        authorized.GET("/user/blacklist", userHandler.UserBlackListRetriveHandler)
        authorized.DELETE("/user/blacklist", userHandler.UserBlackListRemoveHandler)

        // 团队
        authorized.POST("/team", teamHandler.CreateTeamHandler)
        authorized.GET("/team", teamHandler.RetriveTeamHandler)
        authorized.PUT("/team", teamHandler.UpdateTeamHandler)
        authorized.DELETE("/team", teamHandler.DeleteTeamHandler)

        // 角色
        authorized.POST("/role", roleHandler.CreateRoleHandler)
        authorized.GET("/role", roleHandler.RetriveRoleHandler)
        authorized.PUT("/role", roleHandler.UpdateRoleHandler)
        authorized.DELETE("/role", roleHandler.DeleteRoleHandler)
    }

    // 不使用认证中间件
    {
        unauthorized.POST("/user/register", userHandler.CreateUserHandler)
        unauthorized.POST("/user/login", userHandler.UserLoginHandler)

    }

    router.Run(common.Conf.App_host + ":" + common.Conf.App_port)
}
