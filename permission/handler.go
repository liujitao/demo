package permission

import (
    "context"
    "demo/common"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

/*
结构体类型
*/
type PermissionHandler struct {
    ctx        context.Context
    collection *mongo.Collection
}

/*
构造方法
*/
func NewRouteHandler(ctx context.Context, collection *mongo.Collection) *PermissionHandler {
    return &PermissionHandler{
        ctx:        ctx,
        collection: collection,
    }
}

/*
获取路由
*/
func (handler *PermissionHandler) RetriveRouteHandler(c *gin.Context) {
    var response common.Response

    // 查找数据库
    filter := bson.M{}
    cursor, err := handler.collection.Find(handler.ctx, filter)
    if err != nil {
        log.Fatal(err)
    }

    var routes []bson.M
    if err = cursor.All(handler.ctx, &routes); err != nil {
        log.Fatal(err)
    }

    // 输出output
    response.Code = 20000
    response.Message = common.Status[response.Code]
    response.Data = routes
    c.JSON(http.StatusOK, response)
}
