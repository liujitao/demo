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
    Message string      `json:"message"`
    Error   string      `json:"error"`
    Data    interface{} `json:"data"`
}

// 定义返回状态码

/*
1-2位 模块分类 3-4位 模块功能 5-6位 执行状态
用户登录成功 100000
用户登录参数无效 100001
用户名无效 100002
用户密码无效 100003
用户已被锁定 100004
用户token无效 100005

用户退出成功 100100
用户退出失败 100101
用户刷新成功 100300
用户刷新失败 100301
用户改密成功 100300
用户改密失败 100301
用户建立成功 100400
用户建立失败 100401
用户获取成功 100500
用户获取失败 100501
用户更新成功 100600
用户更新失败 100601
用户删除成功 100700
用户删除失败 100701
黑名单获取成功 100800
黑名单获取失败 100801
黑名单增加成功 100900
黑名单增加失败 100901
黑名单移除成功 101000
黑名单移除失败 101001
*/
