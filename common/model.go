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

// 定义输出结果
type List struct {
    Total      int64       `json:"total"`      // 总记录数
    Page_total int64       `json:"page_total"` // 总页数
    Page_index int64       `json:"page_index"` // 当前页号
    Page_size  int64       `json:"page_size"`  // 每页记录数
    Rows       interface{} `json:"rows"`       // 记录内容
}
