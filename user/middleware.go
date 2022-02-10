package user

import (
    "demo/common"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt"
)

const JWT_SECRET = "7ffc6cc0-6dff-11ec-b5e4-97162d942513"

/*
生成token
*/
func GenerateTokens(uuid string) (map[string]string, error) {
    accessExp := time.Now().Add(time.Minute * 5).Unix()
    refreshExp := time.Now().Add(time.Hour * 24 * 7).Unix()

    // access token
    accessToken := jwt.New(jwt.SigningMethodHS256)
    accessClaims := accessToken.Claims.(jwt.MapClaims)
    accessClaims["uuid"] = uuid
    accessClaims["access_exp"] = accessExp
    accessClaims["admin"] = true

    accessString, err := accessToken.SignedString([]byte(JWT_SECRET))
    if err != nil {
        return nil, err
    }

    // refresh token
    refreshToken := jwt.New(jwt.SigningMethodHS256)
    refreshClaims := refreshToken.Claims.(jwt.MapClaims)
    refreshClaims["uuid"] = uuid
    refreshClaims["access_exp"] = accessExp
    refreshClaims["refresh_exp"] = refreshExp

    refreshString, err := refreshToken.SignedString([]byte(JWT_SECRET))
    if err != nil {
        return nil, err
    }

    // return result
    token := map[string]string{"access_token": accessString, "refresh_token": refreshString}
    return token, nil
}

/*
检查token
*/
func VerifyToken(tokenString string) (jwt.MapClaims, bool) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(JWT_SECRET), nil
    })
    if err != nil {
        return nil, false
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok && !token.Valid {
        return claims, false
    }
    return claims, true
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
            response.Code = 000102
            response.Message = common.Status[response.Code]
            c.JSON(http.StatusBadRequest, response)
            c.Abort()
        }

        // 检查token有效性
        claims, ok := VerifyToken(tokenString)
        if !ok {
            response.Code = 100604
            response.Message = common.Status[response.Code]
            c.JSON(http.StatusUnauthorized, response)
            c.Abort()
        }
        uuid := fmt.Sprintf("%v", claims["uuid"])

        // 检查token白名单
        keys, _, err := handler.redisClient.Scan(handler.ctx, 0, uuid+"*", 2).Result()
        if (err != nil) || (len(keys) == 0) {
            response.Code = 100604
            response.Message = common.Status[response.Code]
            c.JSON(http.StatusUnauthorized, response)
            c.Abort()
        }

        // 检查用户黑名单
        if handler.redisClient.SIsMember(handler.ctx, "userblacklist", uuid).Val() {
            response.Code = 100603
            response.Message = common.Status[response.Code]
            c.JSON(http.StatusUnauthorized, response)
            c.Abort()
        }

        c.Next()
    }
}
