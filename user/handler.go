package user

import (
    "context"
    "demo/common"
    "fmt"
    "math"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "github.com/google/uuid"
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
用户注册 & 建立用户
*/
func (handler *UserHandler) CreateUserHandler(c *gin.Context) {
    var response common.Response
    var user UserModel

    // 请求参数parameter
    if err := c.ShouldBindJSON(&user); err != nil {
        response.Code = 000101
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 数据库处理mongo
    user.ID = primitive.NewObjectID()
    user.UUID = uuid.NewString()
    user.Password = common.SetPassword(user.Password)
    user.CreateAt = time.Now()
    user.UpdateAt = time.Now()

    _, err := handler.collection.InsertOne(handler.ctx, user)
    if err != nil {
        response.Code = 000201
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    // 输出output
    response.Code = 100100
    response.Message = common.Status[response.Code]
    response.Data = user
    c.JSON(http.StatusOK, response)
}

/*
获取用户
*/
func (handler *UserHandler) RetriveUserHandler(c *gin.Context) {
    var response common.Response
    var users []UserModel

    // 请求参数parameter
    uuid := c.Query("uuid")
    if uuid == "" {
        response.Code = 000102
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 聚合查询aggregate
    // https://docs.mongodb.com/manual/reference/operator/aggregation/lookup/#std-label-lookup-multiple-joins
    filter := bson.M{"uuid": uuid}
    matchStage := bson.D{{"$match", filter}}

    lookupStage1 := bson.D{{"$lookup", bson.D{{"from", "team"}, {"localField", "uuid"}, {"foreignField", "user_uuid"}, {"as", "team"}}}}
    unwindStage1 := bson.D{{"$unwind", bson.D{{"path", "$team"}, {"preserveNullAndEmptyArrays", true}}}}
    replaceWithStage1 := bson.D{{"$replaceWith", bson.D{{"$mergeObjects", bson.A{bson.D{{"team_name", "$team.team_name"}}, "$$ROOT"}}}}}

    lookupStage2 := bson.D{{"$lookup", bson.D{{"from", "role"}, {"localField", "uuid"}, {"foreignField", "user_uuid"}, {"as", "role"}}}}
    unwindStage2 := bson.D{{"$unwind", bson.D{{"path", "$role"}, {"preserveNullAndEmptyArrays", true}}}}
    replaceWithStage2 := bson.D{{"$replaceWith", bson.D{{"$mergeObjects", bson.A{bson.D{{"role_name", "$role.role_name"}}, "$$ROOT"}}}}}

    projectStage := bson.D{{"$project", bson.D{{"team", 0}, {"role", 0}}}}

    pipeline := mongo.Pipeline{matchStage, lookupStage1, unwindStage1, replaceWithStage1, lookupStage2, unwindStage2, replaceWithStage2, projectStage}
    options := options.Aggregate().SetMaxTime(2 * time.Second)

    cursor, err := handler.collection.Aggregate(handler.ctx, pipeline, options)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    defer cursor.Close(handler.ctx)

    if err = cursor.All(handler.ctx, &users); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 输出output
    response.Code = 100200
    response.Message = common.Status[response.Code]
    response.Data = users[0]
    c.JSON(http.StatusOK, response)
}

/*
更新用户
*/
func (handler *UserHandler) UpdateUserHandler(c *gin.Context) {
    var response common.Response
    var user UserModel

    // 请求参数parameter
    if err := c.ShouldBindJSON(&user); err != nil {
        response.Code = 000101
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusBadRequest, response)
    }

    // 数据库处理mongo
    filter := bson.M{"uuid": user.UUID}
    options := options.FindOneAndUpdate().SetReturnDocument(1)
    update := bson.M{
        "$set": bson.M{
            "user_name": user.UserName,
            "real_name": user.RealName,
            "mobile":    user.Mobile,
            "email":     user.Email,
            "update_at": time.Now(),
        },
    }

    result := handler.collection.FindOneAndUpdate(handler.ctx, filter, update, options)
    if result == nil {
        response.Code = 000203
        response.Message = common.Status[response.Code]
        response.Error = result.Err().Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }
    result.Decode(&user)

    // 输出output
    response.Code = 100300
    response.Message = common.Status[response.Code]
    response.Data = user
    c.JSON(http.StatusOK, response)
}

/*
删除用户
*/
func (handler *UserHandler) DeleteUserHandler(c *gin.Context) {
    var response common.Response

    // 请求参数paramater
    uuid := c.QueryArray("uuid")
    if len(uuid) == 0 {
        response.Code = 000102
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 禁止删除当前登录用户

    // 写入mongo
    filter := bson.M{"uuid": bson.M{"$in": uuid}}
    result, err := handler.collection.DeleteMany(handler.ctx, filter)
    if err != nil {
        response.Code = 000204
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    // 写入redis token  删除access_token refresh_token

    // 输出output
    response.Code = 100400
    response.Message = common.Status[response.Code]
    response.Data = result
    c.JSON(http.StatusOK, response)
}

/*
获取用户列表
*/
func (handler *UserHandler) RetriveUserListHandler(c *gin.Context) {
    var response common.Response
    var users []UserModel

    // 条件匹配match
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

    // 分页pagination
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

    // 排序sort
    var sorts []bson.E
    for filed, order := range c.QueryMap("sort") {
        order, _ := strconv.ParseInt(order, 10, 64)
        sorts = append(sorts, bson.E{filed, order})
    }
    sortStage := bson.D{{"$sort", sorts}}

    // 聚合查询aggregate
    lookupStage1 := bson.D{{"$lookup", bson.D{{"from", "team"}, {"localField", "uuid"}, {"foreignField", "user_uuid"}, {"as", "team"}}}}
    unwindStage1 := bson.D{{"$unwind", bson.D{{"path", "$team"}, {"preserveNullAndEmptyArrays", true}}}}
    replaceWithStage1 := bson.D{{"$replaceWith", bson.D{{"$mergeObjects", bson.A{bson.D{{"team_name", "$team.team_name"}}, "$$ROOT"}}}}}

    lookupStage2 := bson.D{{"$lookup", bson.D{{"from", "role"}, {"localField", "uuid"}, {"foreignField", "user_uuid"}, {"as", "role"}}}}
    unwindStage2 := bson.D{{"$unwind", bson.D{{"path", "$role"}, {"preserveNullAndEmptyArrays", true}}}}
    replaceWithStage2 := bson.D{{"$replaceWith", bson.D{{"$mergeObjects", bson.A{bson.D{{"role_name", "$role.role_name"}}, "$$ROOT"}}}}}

    projectStage := bson.D{{"$project", bson.D{{"team", 0}, {"role", 0}}}}

    pipeline := mongo.Pipeline{matchStage, lookupStage1, unwindStage1, replaceWithStage1, lookupStage2, unwindStage2, replaceWithStage2, projectStage, skipStage, limitStage, sortStage}
    options := options.Aggregate().SetMaxTime(2 * time.Second)

    cursor, err := handler.collection.Aggregate(handler.ctx, pipeline, options)
    if err != nil {
        response.Code = 000202
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }
    defer cursor.Close(handler.ctx)

    if err = cursor.All(handler.ctx, &users); err != nil {
        response.Code = 000202
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    list := &common.List{
        Total:      total,
        Page_total: pageTotal,
        Page_index: pageIndex,
        Page_size:  pageSize,
        Rows:       users,
    }

    // 输出output
    response.Code = 100500
    response.Message = common.Status[response.Code]
    response.Data = list
    c.JSON(http.StatusOK, response)
}

/*
用户登录
*/
func (handler *UserHandler) UserLoginHandler(c *gin.Context) {
    var response common.Response
    var userLogin UserLoginModel

    // 请求参数parameter
    if err := c.ShouldBindJSON(&userLogin); err != nil {
        response.Code = 000101
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 数据库处理mongo （校验密码和锁定）
    var user UserModel
    result := handler.collection.FindOne(handler.ctx, bson.M{"user_name": userLogin.UserName})
    if result == nil {
        response.Code = 100601
        response.Message = common.Status[response.Code]
        response.Error = result.Err().Error()
        c.JSON(http.StatusUnauthorized, response)
        return
    }
    result.Decode(&user)

    // 校验用户密码
    if err := common.VerifyPassword(user.Password, userLogin.Password); err != nil {
        response.Code = 100602
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusUnauthorized, response)
        return
    }

    uuid := user.UUID
    // 检查用户黑名单
    if handler.redisClient.SIsMember(handler.ctx, "userblacklist", uuid).Val() {
        response.Code = 100603
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusUnauthorized, response)
        return
    }

    // 生成token
    tokens, _ := GenerateTokens(uuid)

    // redis处理 (增加新token)
    handler.redisClient.Set(handler.ctx, uuid+"_"+tokens["access_token"], uuid, time.Minute*5)
    handler.redisClient.Set(handler.ctx, uuid+"_"+tokens["refresh_token"], uuid, time.Hour*24*7)

    // 输出output
    response.Code = 100600
    response.Message = common.Status[response.Code]
    response.Data = tokens
    c.JSON(http.StatusOK, response)
}

/*
用户退出
*/
func (handler *UserHandler) UserLogoutHandler(c *gin.Context) {
    var response common.Response

    // 请求参数parameter
    TokenString := c.GetHeader("Authorization")
    if TokenString == "" {
        response.Code = 000102
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 检查token有效性
    claims, ok := VerifyToken(TokenString)
    if !ok {
        response.Code = 100604
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusUnauthorized, response)
        return
    }
    uuid := fmt.Sprintf("%v", claims["uuid"])

    // 检查token白名单
    keys, _, err := handler.redisClient.Scan(handler.ctx, 0, uuid+"*", 2).Result()
    if (err != nil) || (len(keys) == 0) {
        response.Code = 100604
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusUnauthorized, response)
        return
    }

    // redis处理(删除旧access token)
    handler.redisClient.Del(handler.ctx, uuid+"_"+TokenString)

    // 输出output
    response.Code = 100700
    response.Message = common.Status[response.Code]
    c.JSON(http.StatusOK, response)
}

/*
用户刷新
*/
func (handler *UserHandler) UserRefreshHandler(c *gin.Context) {
    var response common.Response

    // 请求参数parameter
    refreshTokenString := c.GetHeader("Authorization")
    if refreshTokenString == "" {
        response.Code = 000102
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 检查token有效性
    claims, ok := VerifyToken(refreshTokenString)
    if !ok {
        response.Code = 100604
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusUnauthorized, response)
        return
    }
    uuid := fmt.Sprintf("%v", claims["uuid"])
    access_exp := int64(claims["access_exp"].(float64))

    // 检查token白名单
    keys, _, err := handler.redisClient.Scan(handler.ctx, 0, uuid+"*", 2).Result()
    if (err != nil) || (len(keys) == 0) {
        response.Code = 100604
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusUnauthorized, response)
        return
    }

    // 检查用户黑名单
    if handler.redisClient.SIsMember(handler.ctx, "userblacklist", uuid).Val() {
        response.Code = 100603
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusUnauthorized, response)
        return
    }

    // 检查刷新时间(accessToken必须过期30秒)
    if time.Now().Sub(time.Unix(access_exp, 0)).Seconds() < 30 {
        response.Code = 100605
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusUnauthorized, response)
        return
    }

    // 生成token
    tokens, _ := GenerateTokens(uuid)

    // redis处理(删除旧refresh token，增加新token)
    handler.redisClient.Del(handler.ctx, uuid+"_"+refreshTokenString)
    handler.redisClient.Set(handler.ctx, uuid+"_"+tokens["access_token"], uuid, time.Minute*5)
    handler.redisClient.Set(handler.ctx, uuid+"_"+tokens["refresh_token"], uuid, time.Hour*24*7)

    // 输出output
    response.Code = 100800
    response.Message = common.Status[response.Code]
    response.Data = tokens
    c.JSON(http.StatusOK, response)
}

/*
用户变更密码
*/
func (handler *UserHandler) UserChanegePasswordHandler(c *gin.Context) {
    var response common.Response
    var userPassword UserPasswordModel

    // 请求参数paramater
    if err := c.ShouldBindJSON(&userPassword); err != nil {
        response.Code = 000101
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 查找数据库
    var user UserModel
    filter := bson.M{"uuid": userPassword.UUID}
    if result := handler.collection.FindOne(handler.ctx, filter); result != nil {
        result.Decode(&user)
    }

    // 校验密码
    if err := common.VerifyPassword(user.Password, userPassword.OldPassword); err != nil {
        response.Code = 100602
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusUnauthorized, response)
        return
    }

    // 写入mongo
    update := bson.M{
        "$set": bson.M{
            "password":  common.SetPassword(userPassword.NewPassword),
            "update_at": time.Now(),
        },
    }

    result := handler.collection.FindOneAndUpdate(handler.ctx, filter, update)
    if result == nil {
        response.Code = 000203
        response.Message = common.Status[response.Code]
        response.Error = result.Err().Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    uuid := user.UUID

    // 生成token
    tokens, _ := GenerateTokens(uuid)

    // redis处理 (删除旧token，增加新token)
    keys, _, _ := handler.redisClient.Scan(handler.ctx, 0, uuid+"*", 2).Result()
    for _, key := range keys {
        handler.redisClient.Del(handler.ctx, key)
    }

    handler.redisClient.Set(handler.ctx, uuid+"_"+tokens["access_token"], uuid, time.Minute*5)
    handler.redisClient.Set(handler.ctx, uuid+"_"+tokens["refresh_token"], uuid, time.Hour*24*7)

    // 输出output
    response.Code = 100900
    response.Message = common.Status[response.Code]
    response.Data = tokens
    c.JSON(http.StatusOK, response)
}

/*
加入用户黑名单
*/
func (handler *UserHandler) UserBlackListAddHandler(c *gin.Context) {
    var response common.Response

    // 请求参数paramater
    uuid := c.QueryArray("uuid")
    if len(uuid) == 0 {
        response.Code = 000102
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 禁止加黑当前登录用户

    // redis
    _, err := handler.redisClient.SAdd(handler.ctx, "userblacklist", uuid).Result()
    if err == redis.Nil {
        response.Code = 000301
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    // 输出output
    response.Code = 101000
    response.Message = common.Status[response.Code]
    c.JSON(http.StatusOK, response)
}

/*
获取用户黑名单
*/
func (handler *UserHandler) UserBlackListRetriveHandler(c *gin.Context) {
    var response common.Response

    // redis
    _, err := handler.redisClient.SMembers(handler.ctx, "userblacklist").Result()
    if err == redis.Nil {
        response.Code = 000302
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    // 输出output
    response.Code = 101100
    response.Message = common.Status[response.Code]
    c.JSON(http.StatusOK, response)
}

/*
移除用户黑名单
*/
func (handler *UserHandler) UserBlackListRemoveHandler(c *gin.Context) {
    var response common.Response

    // 请求参数paramater
    uuid := c.QueryArray("uuid")
    if len(uuid) == 0 {
        response.Code = 000102
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // redis
    _, err := handler.redisClient.SRem(handler.ctx, "userblacklist", uuid).Result()
    if err == redis.Nil {
        response.Code = 000304
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    // 输出output
    response.Code = 101200
    response.Message = common.Status[response.Code]
    c.JSON(http.StatusOK, response)
}

/*
在线用户
*/
func (handler *UserHandler) UserOnlineHandler(c *gin.Context) {
    return
}
