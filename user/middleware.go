package user

import (
    "fmt"
    "net/http"
    "time"

    "github.com/dgrijalva/jwt-go"
    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
)

const JWT_SECRET = "7ffc6cc0-6dff-11ec-b5e4-97162d942513"

type Token struct {
    AccessToken     string `json:"access_token"`
    RefreshToken    string `json:"refresh_token"`
    AccessTokenExp  int64  `json:"access_token_exp"`
    RefreshTokenExp int64  `json:"refresh_token_exp"`
}

/*
生成token
*/
func GenerateTokens(userID string) (interface{}, error) {
    accessExp := time.Now().Add(time.Minute * 5).Unix()
    refreshExp := time.Now().Add(time.Hour * 24 * 7).Unix()

    // access token
    accessToken := jwt.New(jwt.SigningMethodHS256)
    accessClaims := accessToken.Claims.(jwt.MapClaims)
    accessClaims["user_id"] = userID
    accessClaims["admin"] = true
    accessClaims["exp"] = accessExp

    accessString, err := accessToken.SignedString([]byte(JWT_SECRET))
    if err != nil {
        return nil, err
    }

    // refresh token
    refreshToken := jwt.New(jwt.SigningMethodHS256)
    refreshClaims := refreshToken.Claims.(jwt.MapClaims)
    refreshClaims["exp"] = refreshExp

    refreshString, err := refreshToken.SignedString([]byte(JWT_SECRET))
    if err != nil {
        return nil, err
    }

    // return result
    token := Token{
        AccessToken:     accessString,
        RefreshToken:    refreshString,
        AccessTokenExp:  accessExp,
        RefreshTokenExp: refreshExp,
    }
    return token, nil
}

/*
获取token ID
*/
func GetAccessTokenID(tokenString string) interface{} {
    token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(JWT_SECRET), nil
    })

    claims, _ := token.Claims.(jwt.MapClaims)

    return claims["user_id"]
}

/*
认证中间件
*/
func (handler *UserHandler) AuthMiddleWare() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 参数
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"message": "Token has not found"})
            c.Abort()
            return
        }

        // 读redis (black list)
        if _, err := handler.redisClient.Get(handler.ctx, tokenString).Result(); err != redis.Nil {
            c.JSON(http.StatusUnauthorized, gin.H{"message": "Token has been in black list"})
            c.Abort()
            return
        }

        // 解析token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(JWT_SECRET), nil
        })

        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            c.Abort()
            return
        }

        //claims, ok := token.Claims.(jwt.MapClaims)
        _, ok := token.Claims.(jwt.MapClaims)
        if !ok && !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            c.Abort()
            return
        }

        c.Next()
    }
}
