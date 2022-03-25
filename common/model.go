package common

// 定义配置文件
type Configure struct {
    Mongo_host       string
    Mongo_port       string
    Mongo_user       string
    Mongo_password   string
    Mongo_database   string
    Redis_host       string
    Redis_port       string
    Redis_password   string
    Redis_expire_at  int64
    App_host         string
    App_port         string
    App_release_mode bool
    Page_index       int64
    Page_size        int64
    Default_password string
}

// 定义列表输出结果
type List struct {
    Total      int64       `json:"total"`      // 总记录数
    Page_total int64       `json:"page_total"` // 总页数
    Page_index int64       `json:"page_index"` // 当前页号
    Page_size  int64       `json:"page_size"`  // 每页记录数
    Rows       interface{} `json:"rows"`       // 记录内容
}

// 定义输出结果
type Response struct {
    Code    int64       `json:"code"`
    Message string      `json:"message,omitempty"`
    Error   string      `json:"error,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}

/*
   定义返回状态码: 1-2位 模块分类; 3-5位 模块功能
*/

var Status = map[int64]string{
    // 请求处理
    01001: "Request parameter is invalid.",
    01002: "Request parameter is null.",

    // mongo处理
    02001: "Mongodb could not insert data.",
    02002: "Mongodb could not retrieve data.",
    02003: "Mongodb could not update data.",
    02004: "Mongodb could not delete data.",

    // redis处理
    03001: "Redis could not insert data.",
    03002: "Redis could not retrieve data.",
    03003: "Redis could not update data.",
    03004: "Redis could not delete data.",

    // 用户验证
    10000: "User has been successfully login.",
    10001: "User has been successfully logout.",
    10002: "User loginID or password is invalid.",
    10003: "User token has been expired or invalid.",
    10004: "User token wait to expire",
    10005: "User token has been successfully reflushed.",
    10006: "User password has been successfully changed.",
    10007: "User has been locked.",
    // 锁定用户： 存储在redis中的access_token移除；用户id放入redis，refresh_token刷新时比对用户id

    // 用户处理
    11001: "User data has been successfully created.",
    11002: "User data has been successfully retrieved.",
    11003: "User data has been successfully updated.",
    11004: "User data has been successfully deleted.",
    11005: "User list data has been successfully retrieved.",

    // 团队处理
    12001: "Team data has been successfully created.",
    12002: "Team data has been successfully retrieved.",
    12003: "Team data has been successfully updated.",
    12004: "Team data has been successfully deleted.",
    12005: "Team list data has been successfully retrieved.",

    // 角色处理
    13001: "Role data has been successfully created.",
    13002: "Role data has been successfully retrieved.",
    13003: "Role data has been successfully updated.",
    13004: "Role data has been successfully deleted.",
    13005: "Role list data has been successfully retrieved.",
}
