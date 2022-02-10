package role

import (
    "context"
    "demo/common"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
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
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 写入mongo
    role.ID = primitive.NewObjectID()
    role.CreateAt = time.Now()
    role.UpdateAt = time.Now()

    _, err := handler.collection.InsertOne(handler.ctx, role)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new role"})
        return
    }

    // 输出output
    response.Code = 300400
    response.Message = "the role has been create successed"
    response.Data = role
    c.JSON(http.StatusOK, response)
}

/*
获取角色
*/
func (handler *RoleHandler) RetriveRoleHandler(c *gin.Context) {
    var response common.Response
    var role RoleModel

    // 请求参数parameter
    id := c.Query("_id")
    if id == "" {
        c.JSON(http.StatusBadRequest, gin.H{"message": "_id is null"})
        return
    }

    // 数据库处理
    _id, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        c.JSON(http.StatusBadRequest, err.Error())
        return
    }

    filter := bson.M{"_id": _id}
    result := handler.collection.FindOne(handler.ctx, filter)
    if result == nil {
        response.Code = 300201
        response.Message = "the user has been not found"
        response.Error = result.Err().Error()
        c.JSON(http.StatusBadRequest, response)
        return
    }
    result.Decode(&role)

    // 输出output
    response.Code = 300200
    response.Message = "the role has been get successed"
    response.Data = role
    c.JSON(http.StatusOK, response)
}

/*
更新角色
*/
func (handler *RoleHandler) UpdateRoleHandler(c *gin.Context) {
    var response common.Response
    var role RoleModel

    // 请求参数parameter
    id := c.Query("_id")
    if id == "" {
        c.JSON(http.StatusBadRequest, gin.H{"message": "_id is null"})
        return
    }

    if err := c.ShouldBindJSON(&role); err != nil {
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
            "role_name": role.RoleName,
            "describe":  role.Describe,
            "update_at": time.Now(),
        },
    }

    result := handler.collection.FindOneAndUpdate(handler.ctx, filter, update, options)
    if result == nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": result.Err().Error()})
        return
    }
    result.Decode(&role)

    // 输出output
    response.Code = 300300
    response.Message = "the role has been update successed"
    response.Data = role
    c.JSON(http.StatusOK, response)
}

/*
删除角色
*/
func (handler *RoleHandler) DeleteRoleHandler(c *gin.Context) {
    var response common.Response

    // 请求参数paramater
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
    response.Code = 300400
    response.Message = "the role has been delete successed"
    response.Data = result
    c.JSON(http.StatusOK, response)
}
