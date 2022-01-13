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
    // 参数parameter
    var user UserModel
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 写入mongo
    user.ID = primitive.NewObjectID()
    user.Password = common.SetPassword(user.Password)
    user.CreateAt = time.Now()
    user.UpdateAt = time.Now()

    _, err := handler.collection.InsertOne(handler.ctx, user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new user"})
        return
    }

    // 输出output
    c.JSON(http.StatusCreated, user)
}

/*
获取用户
*/
func (handler *UserHandler) RetriveUserHandler(c *gin.Context) {
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
    pipeline := mongo.Pipeline{matchStage, skipStage, limitStage, sortStage}
    options := options.Aggregate().SetMaxTime(2 * time.Second)

    cursor, err := handler.collection.Aggregate(handler.ctx, pipeline, options)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    defer cursor.Close(handler.ctx)

    var users []UserModel
    if err = cursor.All(handler.ctx, &users); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 输出output
    list := &common.List{
        Total:      total,
        Page_total: pageTotal,
        Page_index: pageIndex,
        Page_size:  pageSize,
        Rows:       users,
    }
    c.JSON(http.StatusOK, list)
}

/*
更新用户
*/
func (handler *UserHandler) UpdateUserHandler(c *gin.Context) {
    // 参数parameter
    id := c.Query("_id")
    if id == "" {
        c.JSON(http.StatusBadRequest, gin.H{"message": "_id is null"})
        return
    }

    var user UserModel
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 写入mongo
    _id, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        c.JSON(http.StatusBadRequest, err.Error())
        return
    }

    filter := bson.M{"_id": _id}
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
        c.JSON(http.StatusBadRequest, gin.H{"message": result.Err().Error()})
        return
    }
    result.Decode(&user)

    // 输出output
    c.JSON(http.StatusOK, user)
}

/*
删除用户
*/
func (handler *UserHandler) DeleteUserHandler(c *gin.Context) {
    // 参数paramater
    _id := c.QueryArray("_id")
    if len(_id) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"message": "_id is null"})
        return
    }

    // 禁止删除当前登录用户

    // 写入mongo
    filter := bson.M{"_id": bson.M{"$in": _id}}
    result, err := handler.collection.DeleteMany(handler.ctx, filter)
    if err != nil {
        c.JSON(http.StatusBadRequest, err.Error())
        return
    }

    // 写入redis token  删除access_token refresh_token

    // 输出output
    c.JSON(http.StatusOK, result)
}

/*
变更密码
*/
func (handler *UserHandler) UserChanegePasswordHandler(c *gin.Context) {
    var userPassword UserPasswordModel
    if err := c.ShouldBindJSON(&userPassword); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 查找数据库
    var user UserModel
    filter := bson.M{"_id": userPassword.ID}
    if result := handler.collection.FindOne(handler.ctx, filter); result != nil {
        result.Decode(&user)
    }

    // 校验密码
    if err := common.VerifyPassword(user.Password, userPassword.OldPassword); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "user old password miss"})
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
        c.JSON(http.StatusBadRequest, gin.H{"message": result.Err().Error()})
        return
    }

    // 生成token
    _id := user.ID.Hex()
    tokens, err := GenerateTokens(_id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // redis处理 (删除旧token，增加新token)
    keys, _, err := handler.redisClient.Scan(handler.ctx, 0, _id+"*", 2).Result()
    for _, key := range keys {
        handler.redisClient.Del(handler.ctx, key)
    }

    handler.redisClient.Set(handler.ctx, _id+"_"+tokens["access_token"], _id, time.Minute*5)
    handler.redisClient.Set(handler.ctx, _id+"_"+tokens["refresh_token"], _id, time.Hour*24*7)

    // 输出output
    c.JSON(http.StatusOK, tokens)
}

/*
用户登录
*/
func (handler *UserHandler) UserLoginHandler(c *gin.Context) {
    // 参数parameter
    var response common.Response
    var userLogin UserLoginModel
    if err := c.ShouldBindJSON(&userLogin); err != nil {
        response.Code = 100001
        response.Message = "invalid user login paramater"
        response.Error = err.Error()
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // mongo （校验密码和锁定）
    var user UserModel
    result := handler.collection.FindOne(handler.ctx, bson.M{"user_name": userLogin.UserName})
    if result == nil {
        response.Code = 100002
        response.Message = "invalid user name"
        response.Error = result.Err().Error()
        c.JSON(http.StatusUnauthorized, response)
        return
    }
    result.Decode(&user)

    // 校验用户密码
    if err := common.VerifyPassword(user.Password, userLogin.Password); err != nil {
        response.Code = 100003
        response.Message = "invalid user password"
        response.Error = err.Error()
        c.JSON(http.StatusUnauthorized, response)
        return
    }

    _id := user.ID.Hex()

    // 检查用户黑名单
    if handler.redisClient.SIsMember(handler.ctx, "userblacklist", _id).Val() {
        response.Code = 100004
        response.Message = "the user has been locked"
        c.JSON(http.StatusOK, response)
        return
    }

    // 生成token
    tokens, err := GenerateTokens(_id)
    if err != nil {
        response.Code = 100005
        response.Message = "invalid user token"
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    // redis处理 (增加新token)
    handler.redisClient.Set(handler.ctx, _id+"_"+tokens["access_token"], _id, time.Minute*5)
    handler.redisClient.Set(handler.ctx, _id+"_"+tokens["refresh_token"], _id, time.Hour*24*7)

    // 输出output
    response.Code = 100000
    response.Message = "the user has been login successed"
    response.Data = tokens
    c.JSON(http.StatusOK, response)
}

/*
用户退出
*/
func (handler *UserHandler) UserLogoutHandler(c *gin.Context) {
    // 参数
    TokenString := c.GetHeader("Authorization")
    if TokenString == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "Token not found"})
        return
    }

    // 检查token有效性
    claims, ok := VerifyToken(TokenString)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "token has invaild"})
        return
    }
    _id := fmt.Sprintf("%v", claims["user_id"])

    // 检查token白名单
    keys, _, err := handler.redisClient.Scan(handler.ctx, 0, _id+"*", 2).Result()
    if (err != nil) || (len(keys) == 0) {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "token has invaild"})
        return
    }

    // redis处理(删除旧access token)
    handler.redisClient.Del(handler.ctx, _id+"_"+TokenString)

    // 输出output
    c.JSON(http.StatusOK, gin.H{"message": "user has been logout"})
}

/*
用户刷新
*/
func (handler *UserHandler) UserRefreshHandler(c *gin.Context) {
    // 参数
    refreshTokenString := c.GetHeader("Authorization")
    if refreshTokenString == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "Token not found"})
        return
    }

    // 检查token有效性
    claims, ok := VerifyToken(refreshTokenString)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "token has invaild"})
        return
    }
    _id := fmt.Sprintf("%v", claims["user_id"])
    access_exp := int64(claims["access_exp"].(float64))

    // 检查token白名单
    keys, _, err := handler.redisClient.Scan(handler.ctx, 0, _id+"*", 2).Result()
    if (err != nil) || (len(keys) == 0) {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "token has invaild"})
        return
    }

    // 检查用户黑名单
    if handler.redisClient.SIsMember(handler.ctx, "userblacklist", _id).Val() {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "user has been in black list"})
        return
    }

    // 检查刷新时间(accessToken必须过期30秒)
    if time.Now().Sub(time.Unix(access_exp, 0)).Seconds() < 30 {
        c.JSON(http.StatusOK, gin.H{"message": "token has been not expired"})
        return
    }

    // 生成token
    tokens, err := GenerateTokens(_id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // redis处理(删除旧refresh token，增加新token)
    handler.redisClient.Del(handler.ctx, _id+"_"+refreshTokenString)
    handler.redisClient.Set(handler.ctx, _id+"_"+tokens["access_token"], _id, time.Minute*5)
    handler.redisClient.Set(handler.ctx, _id+"_"+tokens["refresh_token"], _id, time.Hour*24*7)

    // 输出output
    c.JSON(http.StatusOK, tokens)
}

/*
用户黑名单
*/
func (handler *UserHandler) UserBlackListHandler(c *gin.Context) {
    // redis
    result, err := handler.redisClient.SMembers(handler.ctx, "userblacklist").Result()
    if err == redis.Nil {
        c.JSON(http.StatusOK, err)
        return
    }

    // 输出output
    c.JSON(http.StatusOK, result)
}

func (handler *UserHandler) UserBlackListAddHandler(c *gin.Context) {
    // 参数paramater
    _id := c.QueryArray("_id")
    if len(_id) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"message": "_id is null"})
        return
    }

    // 禁止加黑当前登录用户

    // redis
    result, err := handler.redisClient.SAdd(handler.ctx, "userblacklist", _id).Result()
    if err == redis.Nil {
        c.JSON(http.StatusOK, err)
    }

    // 输出output
    c.JSON(http.StatusOK, result)
}

func (handler *UserHandler) UserBlackListRemoveHandler(c *gin.Context) {
    // 参数paramater
    _id := c.QueryArray("_id")
    if len(_id) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"message": "_id is null"})
        return
    }

    // redis
    result, err := handler.redisClient.SRem(handler.ctx, "userblacklist", _id).Result()
    if err == redis.Nil {
        c.JSON(http.StatusOK, err)
        return
    }

    // 输出output
    c.JSON(http.StatusOK, result)
}

/*
在线用户
*/
func (handler *UserHandler) UserOnlineHandler(c *gin.Context) {
    return
}
