package role

import (
    "context"
    "demo/common"
    "math"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

/*
用户类型
*/
type RoleHandler struct {
    ctx        context.Context
    collection *mongo.Collection
}

/*
构造方法
*/
func MewRoleHandler(ctx context.Context, collection *mongo.Collection) *RoleHandler {
    return &RoleHandler{
        ctx:        ctx,
        collection: collection,
    }
}

/*
建立角色
*/
func (handler *RoleHandler) CreateRoleHandler(c *gin.Context) {
    var response common.Response
    var role RoleModel

    // 请求参数parameter
    if err := c.ShouldBindJSON(&role); err != nil {
        response.Code = 000101
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 数据库处理mongo
    role._ID = primitive.NewObjectID()
    role.ID = uuid.NewString()
    role.CreateAt = time.Now()
    role.UpdateAt = time.Now()

    _, err := handler.collection.InsertOne(handler.ctx, role)
    if err != nil {
        response.Code = 000201
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    // 输出output
    response.Code = 120100
    response.Message = common.Status[response.Code]
    response.Data = role
    c.JSON(http.StatusOK, response)
}

/*
获取角色
*/
func (handler *RoleHandler) RetriveRoleHandler(c *gin.Context) {
    var response common.Response
    var roles []RoleModel

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

    lookupStage := bson.D{{"$lookup", bson.D{{"from", "user"}, {"localField", "user_uuid"}, {"foreignField", "uuid"}, {"as", "user"}}}}
    setStage := bson.D{{"$set", bson.D{{"user_name", "$user.real_name"}}}}
    projectStage := bson.D{{"$project", bson.D{{"user", 0}}}}

    pipeline := mongo.Pipeline{matchStage, lookupStage, setStage, projectStage}
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

    if err = cursor.All(handler.ctx, &roles); err != nil {
        response.Code = 000202
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    // 输出output
    response.Code = 120200
    response.Message = common.Status[response.Code]
    response.Data = roles
    c.JSON(http.StatusOK, response)
}

/*
更新角色
*/
func (handler *RoleHandler) UpdateRoleHandler(c *gin.Context) {
    var response common.Response
    var role RoleModel

    // 请求参数parameter
    if err := c.ShouldBindJSON(&role); err != nil {
        response.Code = 000101
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 写入mongo
    // 数据库处理mongo
    filter := bson.M{"id": role.ID}
    options := options.FindOneAndUpdate().SetReturnDocument(1)
    update := bson.M{
        "$set": bson.M{
            "role_name": role.RoleName,
            "describe":  role.Describe,
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
    result.Decode(&role)

    // 输出output
    response.Code = 120300
    response.Message = common.Status[response.Code]
    response.Data = role
    c.JSON(http.StatusOK, response)
}

/*
删除角色
*/
func (handler *RoleHandler) DeleteRoleHandler(c *gin.Context) {
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
    response.Code = 120400
    response.Message = common.Status[response.Code]
    response.Data = result
    c.JSON(http.StatusOK, response)
}

/*
获取角色列表
*/
func (handler *RoleHandler) RetriveRoleListHandler(c *gin.Context) {
    var response common.Response
    var roles []RoleModel

    // 条件匹配match
    filter := bson.M{}
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
    lookupStage := bson.D{{"$lookup", bson.D{{"from", "user"}, {"localField", "user_uuid"}, {"foreignField", "uuid"}, {"as", "user"}}}}
    setStage := bson.D{{"$set", bson.D{{"user_name", "$user.real_name"}}}}
    projectStage := bson.D{{"$project", bson.D{{"user", 0}}}}

    pipeline := mongo.Pipeline{matchStage, lookupStage, setStage, projectStage, skipStage, limitStage, sortStage}
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

    if err = cursor.All(handler.ctx, &roles); err != nil {
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
        Rows:       roles,
    }

    // 输出output
    response.Code = 120500
    response.Message = common.Status[response.Code]
    response.Data = list
    c.JSON(http.StatusOK, response)
}
