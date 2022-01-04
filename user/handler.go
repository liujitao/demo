package user

import (
    "context"
    "demo/common"
    "log"
    "math"
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
    sort := c.QueryMap("sort")

    // match
    //id, _ := primitive.ObjectIDFromHex("61cd47b338e4acbd43065c44")
    filter := bson.M{}
    if search := c.Query("search"); search != "" {
        filter = bson.M{
            "$or": []bson.M{
                {
                    "user_name": bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}},
                },
                {
                    "real_name": bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}},
                },
                {
                    "mobile": bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}},
                },
                {
                    "email": bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}},
                },
                {
                    "create_man": bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}},
                },
                {
                    "update_man": bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}},
                },
            },
        }
    }
    matchStage := bson.D{{"$match", filter}}

    // pagination
    pageIndex, _ := strconv.ParseInt(c.DefaultQuery("pageIndex", "1"), 10, 64)
    pageSize, _ := strconv.ParseInt(c.DefaultQuery("pageSize", "30"), 10, 64)

    total, _ := handler.collection.CountDocuments(handler.ctx, filter)
    pageTotal := int64(math.Ceil(float64(total) / float64(pageSize)))

    limit := pageSize
    skip := limit * (pageIndex - 1)
    if pageIndex > pageTotal {
        skip = 0
        pageIndex = 1
    }

    limitStage := bson.D{{"$limit", limit}}
    skipStage := bson.D{{"$skip", skip}}

    // sort
    var sorts []bson.E
    for filed, order := range sort {
        order, _ := strconv.ParseInt(order, 10, 64)
        sorts = append(sorts, bson.E{filed, order})
    }
    sortStage := bson.D{{"$sort", sorts}}

    // aggregate
    pipeline := mongo.Pipeline{matchStage, skipStage, limitStage, sortStage}
    options := options.Aggregate().SetMaxTime(2 * time.Second)

    cursor, err := handler.collection.Aggregate(handler.ctx, pipeline, options)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    defer cursor.Close(handler.ctx)

    var results []bson.M
    if err = cursor.All(handler.ctx, &results); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // output
    list := &common.List{
        Total:      total,
        Page_total: pageTotal,
        Page_index: pageIndex,
        Page_size:  pageSize,
        Rows:       results,
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
