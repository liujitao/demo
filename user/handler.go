package user

import (
    "context"
    "log"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

/*
用户类型
*/
type UserHandler struct {
    ctx         context.Context
    collection  *mongo.Collection
    redisClient *redis.Client
}

/*
构造方法
*/
func MewUserHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *UserHandler {
    return &UserHandler{
        ctx:         ctx,
        collection:  collection,
        redisClient: redisClient,
    }
}

/*
create user
*/
func (handler *UserHandler) CreateUserHandler(c *gin.Context) {
    var user UserModel
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user.ID = primitive.NewObjectID()
    user.CreateAt = time.Now()
    user.UpdateAt = time.Now()

    _, err := handler.collection.InsertOne(handler.ctx, user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new user"})
        return
    }

    log.Println("Remove data from Redis")
    handler.redisClient.Del("users")

    c.JSON(http.StatusCreated, user)
}

/*
retrieve user
*/
func (handler *UserHandler) RetrieveUserHandler(c *gin.Context) {
    pageIndex := c.DefaultQuery("pageIndex", "10")
    pageSize := c.DefaultQuery("pageSize", "30")
    //sort := c.QueryMap("sort")
    //filter := c.QueryMap("filter")

    index, _ := strconv.ParseInt(pageIndex, 10, 64)
    size, _ := strconv.ParseInt(pageSize, 10, 64)

    // pagination
    limit := size
    skip := limit * (index - 1)

    limitStage := bson.D{{"$limit", limit}}
    skipStage := bson.D{{"$skip", skip}}
    matchStage := bson.D{{"$match", bson.M{}}}

    // pipeline
    pipeline := mongo.Pipeline{limitStage, skipStage, matchStage}
    options := options.Aggregate().SetMaxTime(2 * time.Second)

    // aggregate
    cur, err := handler.collection.Aggregate(handler.ctx, pipeline, options)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    defer cur.Close(handler.ctx)

    var users []interface{}
    for cur.Next(handler.ctx) {
        var document bson.M
        cur.Decode(&document)
        users = append(users, document)
    }

    c.JSON(http.StatusOK, users)
}

/*
update user
*/
func (handler *UserHandler) UpdateUserHandler(c *gin.Context) {
}

/*
delete user
*/
func (handler *UserHandler) DeleteUserHandler(c *gin.Context) {
}
