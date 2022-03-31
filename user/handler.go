package user

import (
    "context"
    "demo/common"
    "math"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "github.com/rs/xid"
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
        response.Code = 21001
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 数据库处理mongo
    user._ID = primitive.NewObjectID()
    user.ID = xid.New().String()
    user.Password = common.SetPassword(user.Password)
    user.CreateAt = time.Now()
    user.UpdateAt = time.Now()

    if _, err := handler.collection.InsertOne(handler.ctx, user); err != nil {
        response.Code = 22001
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    // 输出output
    response.Code = 10000
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
    id := c.Query("id")
    if id == "" {
        response.Code = 21001
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 聚合查询aggregate
    // https://docs.mongodb.com/manual/reference/operator/aggregation/lookup/#std-label-lookup-multiple-joins
    filter := bson.M{"id": id}
    matchStage := bson.D{{"$match", filter}}

    lookupStage1 := bson.D{{"$lookup", bson.D{{"from", "team"}, {"localField", "id"}, {"foreignField", "user_id"}, {"as", "team"}}}}
    unwindStage1 := bson.D{{"$unwind", bson.D{{"path", "$team"}, {"preserveNullAndEmptyArrays", true}}}}
    replaceWithStage1 := bson.D{{"$replaceWith", bson.D{{"$mergeObjects", bson.A{bson.D{{"team_name", "$team.team_name"}}, "$$ROOT"}}}}}

    lookupStage2 := bson.D{{"$lookup", bson.D{{"from", "role"}, {"localField", "id"}, {"foreignField", "user_id"}, {"as", "role"}}}}
    unwindStage2 := bson.D{{"$unwind", bson.D{{"path", "$role"}, {"preserveNullAndEmptyArrays", true}}}}
    replaceWithStage2 := bson.D{{"$replaceWith", bson.D{{"$mergeObjects", bson.A{bson.D{{"role_name", "$role.role_name"}}, "$$ROOT"}}}}}

    projectStage := bson.D{{"$project", bson.D{{"team", 0}, {"role", 0}}}}

    pipeline := mongo.Pipeline{matchStage, lookupStage1, unwindStage1, replaceWithStage1, lookupStage2, unwindStage2, replaceWithStage2, projectStage}
    options := options.Aggregate().SetMaxTime(2 * time.Second)

    cursor, err := handler.collection.Aggregate(handler.ctx, pipeline, options)
    if err != nil {
        response.Code = 22002
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    defer cursor.Close(handler.ctx)

    if err = cursor.All(handler.ctx, &users); err != nil {
        response.Code = 22002
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 输出output
    response.Code = 11002
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
        response.Code = 21001
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusBadRequest, response)
    }

    // 数据库处理mongo
    filter := bson.M{"id": user.ID}
    options := options.FindOneAndUpdate().SetReturnDocument(1)
    update := bson.M{
        "$set": bson.M{
            "user_name": user.UserName,
            "mobile":    user.Mobile,
            "email":     user.Email,
            "update_at": time.Now(),
        },
    }

    if err := handler.collection.FindOneAndUpdate(handler.ctx, filter, update, options).Decode(&user); err != nil {
        response.Code = 22003
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    // 输出output
    response.Code = 11003
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
    id := c.QueryArray("id")
    if len(id) == 0 {
        response.Code = 21001
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 禁止删除当前登录用户

    // 写入mongo
    filter := bson.M{"id": bson.M{"$in": id}}
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
    lookupStage1 := bson.D{{"$lookup", bson.D{{"from", "team"}, {"localField", "id"}, {"foreignField", "user_uuid"}, {"as", "team"}}}}
    unwindStage1 := bson.D{{"$unwind", bson.D{{"path", "$team"}, {"preserveNullAndEmptyArrays", true}}}}
    replaceWithStage1 := bson.D{{"$replaceWith", bson.D{{"$mergeObjects", bson.A{bson.D{{"team_name", "$team.team_name"}}, "$$ROOT"}}}}}

    lookupStage2 := bson.D{{"$lookup", bson.D{{"from", "role"}, {"localField", "id"}, {"foreignField", "user_uuid"}, {"as", "role"}}}}
    unwindStage2 := bson.D{{"$unwind", bson.D{{"path", "$role"}, {"preserveNullAndEmptyArrays", true}}}}
    replaceWithStage2 := bson.D{{"$replaceWith", bson.D{{"$mergeObjects", bson.A{bson.D{{"role_name", "$role.role_name"}}, "$$ROOT"}}}}}

    projectStage := bson.D{{"$project", bson.D{{"team", 0}, {"role", 0}}}}

    pipeline := mongo.Pipeline{matchStage, lookupStage1, unwindStage1, replaceWithStage1, lookupStage2, unwindStage2, replaceWithStage2, projectStage, skipStage, limitStage, sortStage}
    options := options.Aggregate().SetMaxTime(2 * time.Second)

    cursor, err := handler.collection.Aggregate(handler.ctx, pipeline, options)
    if err != nil {
        response.Code = 22002
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }
    defer cursor.Close(handler.ctx)

    if err = cursor.All(handler.ctx, &users); err != nil {
        response.Code = 22002
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
    response.Code = 11005
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
        response.Code = 21001
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 数据库处理mongo （校验密码 & 用户激活）
    var user UserModel
    filter :=
        bson.D{{"$and", bson.A{
            bson.D{{"$or", bson.A{
                bson.D{{"mobile", userLogin.LoginID}},
                bson.D{{"email", userLogin.LoginID}},
            }}},
            bson.D{{"active", 1}},
        }}}

    if err := handler.collection.FindOne(handler.ctx, filter).Decode(&user); err != nil {
        response.Code = 22002
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusUnauthorized, response)
        return
    }

    // 校验用户密码
    if err := common.VerifyPassword(user.Password, userLogin.Password); err != nil {
        response.Code = 22002
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusUnauthorized, response)
        return
    }

    // 生成token
    tokens := GenerateTokens(user.ID)
    tokens["id"] = user.ID

    // redis处理 (写入token)
    handler.redisClient.Set(handler.ctx, user.ID, tokens["access_token"], time.Minute*time.Duration(common.Conf.Access_token_exp))

    // 输出output
    response.Code = 10000
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
        response.Code = 21001
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 从token获得用户id
    id, _ := ParseToken(TokenString)

    // redis处理(移除token)
    handler.redisClient.Del(handler.ctx, id)

    // 输出output
    response.Code = 10001
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
        response.Code = 01001
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 从token获得用户id
    id, _ := ParseToken(refreshTokenString)

    // 生成token
    tokens := GenerateTokens(id)
    tokens["id"] = id

    // redis处理(删除旧refresh token，增加新token)
    handler.redisClient.Del(handler.ctx, id)
    handler.redisClient.Set(handler.ctx, id, tokens["access_token"], time.Minute*time.Duration(common.Conf.Access_token_exp))

    // 输出output
    response.Code = 10000
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
        response.Code = 01001
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 查找数据库
    var user UserModel
    filter := bson.M{"id": userPassword.ID}
    if result := handler.collection.FindOne(handler.ctx, filter); result != nil {
        result.Decode(&user)
    }

    // 校验密码
    if err := common.VerifyPassword(user.Password, userPassword.OldPassword); err != nil {
        response.Code = 10002
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

    // 生成token
    tokens := GenerateTokens(user.ID)
    tokens["id"] = user.ID

    // redis处理 (写入token)
    handler.redisClient.Set(handler.ctx, user.ID, tokens["access_token"], time.Minute*time.Duration(common.Conf.Access_token_exp))

    // 输出output
    response.Code = 10000
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
    id := c.QueryArray("id")
    if len(id) == 0 {
        response.Code = 01001
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 禁止加黑当前登录用户

    // redis处理 (写入UserLock)
    _, err := handler.redisClient.SAdd(handler.ctx, "UserLock", id).Result()
    if err == redis.Nil {
        response.Code = 03001
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    // 输出output
    response.Code = 10000
    response.Message = common.Status[response.Code]
    c.JSON(http.StatusOK, response)
}

/*
获取用户黑名单
*/
func (handler *UserHandler) UserBlackListRetriveHandler(c *gin.Context) {
    var response common.Response

    // redis
    _, err := handler.redisClient.SMembers(handler.ctx, "UserLock").Result()
    if err == redis.Nil {
        response.Code = 03002
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    // 输出output
    response.Code = 10000
    response.Message = common.Status[response.Code]
    c.JSON(http.StatusOK, response)
}

/*
移除用户黑名单
*/
func (handler *UserHandler) UserBlackListRemoveHandler(c *gin.Context) {
    var response common.Response

    // 请求参数paramater
    id := c.QueryArray("id")
    if len(id) == 0 {
        response.Code = 01001
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // redis
    _, err := handler.redisClient.SRem(handler.ctx, "userblacklist", id).Result()
    if err == redis.Nil {
        response.Code = 03004
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    // 输出output
    response.Code = 10000
    response.Message = common.Status[response.Code]
    c.JSON(http.StatusOK, response)
}

/*
在线用户
*/
func (handler *UserHandler) UserOnlineHandler(c *gin.Context) {
    return
}
