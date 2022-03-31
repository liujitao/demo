package user

import (
    "demo/common"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "github.com/golang-jwt/jwt"
)

const JWT_SECRET = "7ffc6cc0-6dff-11ec-b5e4-97162d942513"

/*
生成token
*/
func GenerateTokens(id string) map[string]string {
    // access token
    accessToken := jwt.New(jwt.GetSigningMethod("HS256"))
    accessToken.Claims = jwt.MapClaims{
        "id":  id,
        "exp": time.Now().Add(time.Minute * time.Duration(common.Conf.Access_token_exp)).Unix(),
    }
    accessString, _ := accessToken.SignedString([]byte(JWT_SECRET))

    // refresh token
    refreshToken := jwt.New(jwt.GetSigningMethod("HS256"))
    refreshToken.Claims = jwt.MapClaims{
        "id":  id,
        "exp": time.Now().Add(time.Minute * time.Duration(common.Conf.Refresh_token_exp)).Unix(),
    }
    refreshString, _ := refreshToken.SignedString([]byte(JWT_SECRET))

    // return
    tokens := map[string]string{"access_token": accessString, "refresh_token": refreshString}
    return tokens
}

/*
解析token
*/
func ParseToken(tokenString string) (string, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(JWT_SECRET), nil
    })

    if err != nil {
        return "", nil
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return fmt.Sprintf("%v", claims["id"]), nil
    } else {
        return "", nil
    }
}

/*
认证中间件
*/
func (handler *UserHandler) AuthMiddleWare() gin.HandlerFunc {
    return func(c *gin.Context) {
        var response common.Response

        // 请求参数
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            response.Code = 01001
            response.Message = common.Status[response.Code]
            c.JSON(http.StatusBadRequest, response)
            c.Abort()
        }

        // 检查token有效性
        id, err := ParseToken(tokenString)
        if err != nil {
            response.Code = 10003
            response.Message = common.Status[response.Code]
            c.JSON(http.StatusUnauthorized, response)
            c.Abort()
        }

        // 检查redis是否存在
        if val, err := handler.redisClient.Get(handler.ctx, id).Result(); (err == redis.Nil) || (val != tokenString) {
            response.Code = 10003
            response.Message = common.Status[response.Code]
            c.JSON(http.StatusUnauthorized, response)
            c.Abort()
        }

        // 检查token用户是否在锁定名单
        if handler.redisClient.SIsMember(handler.ctx, "UserLock", id).Val() {
            response.Code = 100607
            response.Message = common.Status[response.Code]
            c.JSON(http.StatusUnauthorized, response)
            c.Abort()
        }

        c.Next()
    }
}
