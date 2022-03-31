package common

import (
    "encoding/json"
    "io/ioutil"
    "log"

    "golang.org/x/crypto/bcrypt"
)

/*
全局配置文件
*/
var Conf *Configure

func GetConfig() {
    content, err := ioutil.ReadFile("common/config.json")
    if err != nil {
        log.Fatal("Error when opening file: ", err)
    }

    err = json.Unmarshal(content, &Conf)
    if err != nil {
        log.Fatal("Error during Unmarshal(): ", err)
    }
}

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
