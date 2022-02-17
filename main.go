package main

import (
    "context"
    "demo/common"
    "demo/role"
    "demo/team"
    "demo/user"
    "encoding/json"
    "io/ioutil"
    "log"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"
)

// 定义变量
var Conf *common.Configure
var userHandler *user.UserHandler
var teamHandler *team.TeamHandler
var roleHandler *role.RoleHandler

// 初始化
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
    teamHandler = team.MewTeamHandler(ctx, db.Collection("team"))
    roleHandler = role.MewRoleHandler(ctx, db.Collection("role"))

    // 导入用户数据
    if count, _ := db.Collection("user").CountDocuments(ctx, bson.M{}); count == 0 {
        content, err := ioutil.ReadFile("common/userData.json")
        if err != nil {
            log.Fatal("Error when opening file: ", err)
        }

        var users []user.UserModel
        err = json.Unmarshal(content, &users)
        if err != nil {
            log.Fatal("Error during Unmarshal(): ", err)
        }

        var document interface{}
        for _, u := range users {
            document = bson.D{
                {"_id", primitive.NewObjectID()},
                {"uuid", u.UUID},
                {"user_name", u.UserName},
                {"real_name", u.RealName},
                {"mobile", u.Mobile},
                {"email", u.Email},
                {"password", common.SetPassword(u.Password)},
                {"create_at", time.Now()},
                {"update_at", time.Now()},
            }
            _, err = db.Collection("user").InsertOne(ctx, document)
            if err != nil {
                log.Fatal("Error insert user data: ", err)
            }
        }
    }

    // 导入团队数据
    if count, _ := db.Collection("team").CountDocuments(ctx, bson.M{}); count == 0 {
        content, err := ioutil.ReadFile("common/teamData.json")
        if err != nil {
            log.Fatal("Error when opening file: ", err)
        }

        var teams []team.TeamModel
        err = json.Unmarshal(content, &teams)
        if err != nil {
            log.Fatal("Error during Unmarshal(): ", err)
        }

        var document interface{}
        for _, t := range teams {
            document = bson.D{
                {"_id", primitive.NewObjectID()},
                {"uuid", t.UUID},
                {"team_name", t.TeamName},
                {"describe", t.Describe},
                {"user_uuid", t.User},
                {"create_at", time.Now()},
                {"update_at", time.Now()},
            }
            _, err = db.Collection("team").InsertOne(ctx, document)
            if err != nil {
                log.Fatal("Error insert team data: ", err)
            }
        }
    }

    // 导入角色数据
    if count, _ := db.Collection("role").CountDocuments(ctx, bson.M{}); count == 0 {
        content, err := ioutil.ReadFile("common/roleData.json")
        if err != nil {
            log.Fatal("Error when opening file: ", err)
        }

        var roles []role.RoleModel
        err = json.Unmarshal(content, &roles)
        if err != nil {
            log.Fatal("Error during Unmarshal(): ", err)
        }

        var document interface{}
        for _, r := range roles {
            document = bson.D{
                {"_id", primitive.NewObjectID()},
                {"uuid", r.UUID},
                {"role_name", r.RoleName},
                {"describe", r.Describe},
                {"user_uuid", r.User},
                {"create_at", time.Now()},
                {"update_at", time.Now()},
            }
            _, err = db.Collection("role").InsertOne(ctx, document)
            if err != nil {
                log.Fatal("Error insert role data: ", err)
            }
        }
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
        authorized.GET("/user/list", userHandler.RetriveUserListHandler)
        authorized.POST("/user/change_password", userHandler.UserChanegePasswordHandler)

        // 用户黑名单
        authorized.POST("/user/blacklist", userHandler.UserBlackListAddHandler)
        authorized.GET("/user/blacklist", userHandler.UserBlackListRetriveHandler)
        authorized.DELETE("/user/blacklist", userHandler.UserBlackListRemoveHandler)

        // 团队
        /*
           authorized.POST("/team", teamHandler.CreateTeamHandler)
           authorized.GET("/team", teamHandler.RetriveTeamHandler)
           authorized.PUT("/team", teamHandler.UpdateTeamHandler)
           authorized.DELETE("/team", teamHandler.DeleteTeamHandler)
        */

        // 角色
        /*
           authorized.POST("/role", roleHandler.CreateRoleHandler)
           authorized.GET("/role", roleHandler.RetriveRoleHandler)
           authorized.PUT("/role", roleHandler.UpdateRoleHandler)
           authorized.DELETE("/role", roleHandler.DeleteRoleHandler)
        */
    }

    // 不使用认证中间件
    {
        router.POST("/user/register", userHandler.CreateUserHandler)
        router.POST("/user/login", userHandler.UserLoginHandler)
        router.GET("/user/logout", userHandler.UserLogoutHandler)
        router.GET("/user/refresh", userHandler.UserRefreshHandler)

        // 临时
        router.POST("/team", teamHandler.CreateTeamHandler)
        router.GET("/team", teamHandler.RetriveTeamHandler)
        router.PUT("/team", teamHandler.UpdateTeamHandler)
        router.DELETE("/team", teamHandler.DeleteTeamHandler)
        router.GET("/team/list", teamHandler.RetriveTeamListHandler)

        // 临时
        router.POST("/role", roleHandler.CreateRoleHandler)
        router.GET("/role", roleHandler.RetriveRoleHandler)
        router.PUT("/role", roleHandler.UpdateRoleHandler)
        router.DELETE("/role", roleHandler.DeleteRoleHandler)
        router.GET("/role/list", roleHandler.RetriveRoleListHandler)
    }

    router.Run(Conf.App_host + ":" + Conf.App_port)
}
