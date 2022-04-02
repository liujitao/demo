package team

import (
    "context"
    "demo/common"
    "math"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/rs/xid"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

/*
结构体类型
*/
type TeamHandler struct {
    ctx        context.Context
    collection *mongo.Collection
}

/*
构造方法
*/
func NewTeamHandler(ctx context.Context, collection *mongo.Collection) *TeamHandler {
    return &TeamHandler{
        ctx:        ctx,
        collection: collection,
    }
}

/*
建立团队
*/
func (handler *TeamHandler) CreateTeamHandler(c *gin.Context) {
    var response common.Response
    var team TeamModel

    // 请求参数parameter
    if err := c.ShouldBindJSON(&team); err != nil {
        response.Code = 000101
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 数据库处理mongo
    team._ID = primitive.NewObjectID()
    team.ID = xid.New().String()
    team.CreateAt = time.Now()
    team.UpdateAt = time.Now()

    _, err := handler.collection.InsertOne(handler.ctx, team)
    if err != nil {
        response.Code = 000201
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    // 输出output
    response.Code = 110100
    response.Message = common.Status[response.Code]
    response.Data = team
    c.JSON(http.StatusOK, response)
}

/*
获取团队
*/
func (handler *TeamHandler) RetriveTeamHandler(c *gin.Context) {
    var response common.Response
    var teams []TeamModel

    // 请求参数parameter
    id := c.Query("id")
    if id == "" {
        response.Code = 000102
        response.Message = common.Status[response.Code]
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 聚合查询aggregate
    // https://docs.mongodb.com/manual/reference/operator/aggregation/lookup/#std-label-lookup-multiple-joins
    filter := bson.M{"id": id}
    matchStage := bson.D{{"$match", filter}}

    lookupStage := bson.D{{"$lookup", bson.D{{"from", "user"}, {"localField", "user_uuid"}, {"foreignField", "id"}, {"as", "user"}}}}
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

    if err = cursor.All(handler.ctx, &teams); err != nil {
        response.Code = 000202
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusInternalServerError, response)
        return
    }

    // 输出output
    response.Code = 110200
    response.Message = common.Status[response.Code]
    response.Data = teams
    c.JSON(http.StatusOK, response)
}

/*
更新团队
*/
func (handler *TeamHandler) UpdateTeamHandler(c *gin.Context) {
    var response common.Response
    var team TeamModel

    // 请求参数parameter
    if err := c.ShouldBindJSON(&team); err != nil {
        response.Code = 000101
        response.Message = common.Status[response.Code]
        response.Error = err.Error()
        c.JSON(http.StatusBadRequest, response)
        return
    }

    // 写入mongo
    // 数据库处理mongo
    filter := bson.M{"id": team.ID}
    options := options.FindOneAndUpdate().SetReturnDocument(1)
    update := bson.M{
        "$set": bson.M{
            "team_name": team.TeamName,
            "describe":  team.Describe,
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
    result.Decode(&team)

    // 输出output
    response.Code = 110300
    response.Message = common.Status[response.Code]
    response.Data = team
    c.JSON(http.StatusOK, response)
}

/*
删除团队
*/
func (handler *TeamHandler) DeleteTeamHandler(c *gin.Context) {
    var response common.Response

    // 请求参数paramater
    id := c.QueryArray("id")
    if len(id) == 0 {
        response.Code = 000102
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
    response.Code = 110400
    response.Message = common.Status[response.Code]
    response.Data = result
    c.JSON(http.StatusOK, response)
}

/*
获取团队列表
*/
func (handler *TeamHandler) RetriveTeamListHandler(c *gin.Context) {
    var response common.Response
    var teams []TeamModel

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
    lookupStage := bson.D{{"$lookup", bson.D{{"from", "user"}, {"localField", "user_uuid"}, {"foreignField", "id"}, {"as", "user"}}}}
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

    if err = cursor.All(handler.ctx, &teams); err != nil {
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
        Rows:       teams,
    }

    // 输出output
    response.Code = 110500
    response.Message = common.Status[response.Code]
    response.Data = list
    c.JSON(http.StatusOK, response)
}

/*
获取团队成员
*/
func (handler *TeamHandler) RetriveTeamMemberHandler(c *gin.Context) {
}
