package user

import (
    "context"
    "demo/common"
    "encoding/json"
    "fmt"
    "math"
    "net/http"
    "strconv"
    "time"

    "github.com/dgrijalva/jwt-go"
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
建立用户
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

    // 写入redis
    data, _ := json.Marshal(user)
    handler.redisClient.Set(handler.ctx, user.ID.Hex(), data, time.Hour)

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

    var documents []bson.M
    if err = cursor.All(handler.ctx, &documents); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 输出output
    list := &common.List{
        Total:      total,
        Page_total: pageTotal,
        Page_index: pageIndex,
        Page_size:  pageSize,
        Rows:       documents,
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

    var document bson.M
    result := handler.collection.FindOneAndUpdate(handler.ctx, filter, update, options)
    if result == nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": result.Err().Error()})
        return
    }
    result.Decode(&document)

    // 写入redis
    data, _ := json.Marshal(document)
    handler.redisClient.Set(handler.ctx, id, data, time.Hour)

    // 输出output
    c.JSON(http.StatusOK, document)
}

/*
删除用户
*/
func (handler *UserHandler) DeleteUserHandler(c *gin.Context) {
    // 参数paramater
    array := c.QueryArray("_id")
    if len(array) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"message": "_id is null"})
        return
    }

    var _id []interface{}
    for _, id := range array {
        i, err := primitive.ObjectIDFromHex(id)
        if err != nil {
            c.JSON(http.StatusBadRequest, err.Error())
            return
        }
        _id = append(_id, i)
    }

    // 写入mongo
    filter := bson.M{"_id": bson.M{"$in": _id}}
    result, err := handler.collection.DeleteMany(handler.ctx, filter)
    if err != nil {
        c.JSON(http.StatusBadRequest, err.Error())
        return
    }

    // 写入redis
    for _, id := range array {
        handler.redisClient.Del(handler.ctx, id)
    }

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

    // 读取数据库
    var document bson.M
    filter := bson.M{"_id": userPassword.ID}
    if result := handler.collection.FindOne(handler.ctx, filter); result != nil {
        result.Decode(&document)
    }

    // 校验密码
    passwordHash := fmt.Sprintf("%v", document["password"])
    if err := common.VerifyPassword(passwordHash, userPassword.OldPassword); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "old password miss"})
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
    result.Decode(&document)

    // 写入redis
    data, _ := json.Marshal(document)
    handler.redisClient.Set(handler.ctx, userPassword.ID.Hex(), data, time.Hour)

    // 输出output
    c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

/*
认证中间件
*/
func AuthMiddleWare() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 参数
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"message": "Token not found"})
            return
        }

        // 解析token
        token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
            return []byte(JWT_SECRET), nil
        })

        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        if _, ok := token.Claims.(*CustomClaims); !ok && !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"message": "Ivaild token"})
            return
        }

        c.Next()
    }
}

/*
用户登录
*/
func (handler *UserHandler) UserLoginHandler(c *gin.Context) {
    // 参数parameter
    var user UserLoginModel
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // mongo
    var document bson.M
    result := handler.collection.FindOne(handler.ctx, bson.M{"user_name": user.UserName})
    if result == nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": result.Err().Error()})
        return
    }
    result.Decode(&document)

    // 校验用户密码
    passwordHash := fmt.Sprintf("%v", document["password"])
    if err := common.VerifyPassword(passwordHash, user.Password); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "invaild password"})
        return
    }

    // 生成JWT token
    expirationTime := time.Now().Add(time.Minute * 15)
    claims := &CustomClaims{
        user.UserName,
        jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(JWT_SECRET))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 输出output
    c.JSON(http.StatusOK, gin.H{"token": tokenString, "expires": expirationTime})
}

/*
用户刷新
*/
func (handler *UserHandler) UserRefreshHandler(c *gin.Context) {
    // 参数
    tokenString := c.GetHeader("Authorization")
    if tokenString == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "Token not found"})
        return
    }

    // 解析token
    token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(JWT_SECRET), nil
    })

    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    claims, ok := token.Claims.(*CustomClaims)
    if !ok && !token.Valid {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "Ivaild token"})
        return
    }

    // 判断过期
    if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > time.Second*30 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Token is not expired yet"})
        return
    }
    // 重新生成JWT token
    expirationTime := time.Now().Add(time.Minute * 15)

    claims.ExpiresAt = expirationTime.Unix()
    newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    newTokenString, err := newToken.SignedString([]byte(JWT_SECRET))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 输出output
    c.JSON(http.StatusOK, gin.H{"token": newTokenString, "expires": expirationTime})
}

/*
用户退出
*/
func (handler *UserHandler) UserLogoutHandler(c *gin.Context) {
}
