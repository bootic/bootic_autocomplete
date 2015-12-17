package lib

import (
	"net/http"
)

type Item struct {
	Links map[string]*Link `json:"_links"`
	Title string           `json:"title"`
	Price float64          `json:"price"`
}

type Link struct {
	Href string `json:"href"`
}

type Results struct {
	Links      map[string]interface{} `json:"_links"`
	TotalItems uint64                 `json:"total_items"`
	Page       uint64                 `json:"page"`
	PerPage    uint64                 `json:"per_page"`
	Embedded   map[string]interface{} `json:"_embedded"`
}

type Context struct {
	Page    uint64
	PerPage uint64
	Q       string
	CdnHost string
	Req     *http.Request
}
