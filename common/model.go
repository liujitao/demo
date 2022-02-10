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
   定义返回状态码: 1-2位 模块分类; 3-4位 模块功能; 5-6位 序列号
*/

var Status = map[int64]string{
    // 请求参数
    000101: "Request parameter is invalid.",
    000102: "Request parameter is null.",

    // mongo处理
    000201: "Mongodb could not insert data.",
    000202: "Mongodb could not retrieve data.",
    000203: "Mongodb could not update data.",
    000204: "Mongodb could not delete data.",

    // redis处理
    000301: "Redis could not insert data.",
    000302: "Redis could not retrieve data.",
    000303: "Redis could not update data.",
    000304: "Redis could not delete data.",

    // 用户
    100100: "User data has been successfully created.",
    100200: "User data has been successfully retrieved.",
    100300: "User data has been successfully updated.",
    100400: "User data has been successfully deleted.",
    100500: "User list data has been successfully retrieved.",

    100600: "User has been successfully login.",
    100601: "User name is invalid.",
    100602: "User password is invalid.",
    100603: "User has been locked.",
    100604: "User token is invalid.",
    100605: "User token wait to expire",

    100700: "User has been successfully logout.",
    100800: "User has been successfully flushed.",
    100900: "User password has been successfully changed.",

    101000: "User blacklist has been successfully added.",
    101100: "User blacklist has been successfully retrieved.",
    101200: "User blacklist has been successfully removed.",

    // 团队

    // 角色
}
