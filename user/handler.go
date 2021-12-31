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
    sort := c.QueryMap("sort")
    //filter := c.QueryMap("filter")

    index, _ := strconv.ParseInt(pageIndex, 10, 64)
    size, _ := strconv.ParseInt(pageSize, 10, 64)

    // pagination
    limit := size
    skip := limit * (index - 1)

    limitStage := bson.D{{"$limit", limit}}
    skipStage := bson.D{{"$skip", skip}}

    // sort
    var sorts []bson.E
    for filed, order := range sort {
        order, _ := strconv.ParseInt(order, 10, 64)
        sorts = append(sorts, bson.E{filed, order})
    }
    sortStage := bson.D{{"$sort", sorts}}

    // match
    //id, _ := primitive.ObjectIDFromHex("61cd47b338e4acbd43065c44")
    matchStage := bson.D{{"$match", bson.M{}}}

    // pipeline
    pipeline := mongo.Pipeline{matchStage, skipStage, limitStage, sortStage}
    options := options.Aggregate().SetMaxTime(2 * time.Second)

    // aggregate
    cursor, err := handler.collection.Aggregate(handler.ctx, pipeline, options)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    defer cursor.Close(handler.ctx)

    var list []bson.M
    if err = cursor.All(handler.ctx, &list); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, list)
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
