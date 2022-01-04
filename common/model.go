package common

import "go.mongodb.org/mongo-driver/bson"

type Configure struct {
	Mongo_host       string
	Mongo_port       string
	Mongo_user       string
	Mongo_password   string
	Mongo_database   string
	Redis_host       string
	Redis_port       string
	Redis_password   string
	App_host         string
	App_port         string
	App_release_mode bool
}

type List struct {
	Total      int64    `json:"total"`      // total records
	Page_total int64    `json:"page_total"` // total pages
	Page_index int64    `json:"page_index"` // current page
	Page_size  int64    `json:"page_size"`  // records per page
	Rows       []bson.M `json:"rows"`       // records
}
