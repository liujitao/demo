package common

import (
    "time"

    "github.com/dgrijalva/jwt-go"
    "golang.org/x/crypto/bcrypt"
)

/*
明文加密
*/
func SetPassword(password string) string {
    bytePassword := []byte(password)
    passwordHash, _ := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
    return string(passwordHash)
}

/*
验证密文
*/
func VerifyPassword(passwordHash string, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
}

/*
生成token
*/
const JWT_SECRET = "7ffc6cc0-6dff-11ec-b5e4-97162d942513"

func GenerateTokens(userID string) (map[string]string, error) {
    // access token
    token := jwt.New(jwt.SigningMethodHS256)
    claims := token.Claims.(jwt.MapClaims)
    claims["user_id"] = userID
    claims["admin"] = true
    claims["exp"] = time.Now().Add(time.Minute * 5).Unix()

    tokenString, err := token.SignedString([]byte(JWT_SECRET))
    if err != nil {
        return nil, err
    }

    // refresh token
    refreshToken := jwt.New(jwt.SigningMethodHS256)
    refreshClaims := refreshToken.Claims.(jwt.MapClaims)
    refreshClaims["exp"] = time.Now().Add(time.Hour * 24).Unix()

    refreshString, err := refreshToken.SignedString([]byte(JWT_SECRET))
    if err != nil {
        return nil, err
    }

    return map[string]string{"access_token": tokenString, "refresh_token": refreshString}, nil
}
