package lib

import (
	"github.com/belogik/goes"
	"net/http"
)

type QueryBuilder interface {
	Build(*Context) map[string]interface{}
}

type Searcher interface {
	Search(map[string]interface{}) (*goes.Response, error)
}

type Presenter interface {
	Present(*goes.Response, *Context) *Results
}

type Item struct {
	Links          map[string]*Link `json:"_links"`
	Title          string           `json:"title"`
	Price          uint64           `json:"price"`
	FormattedPrice string           `json:"formatted_price"`
	Tags           []interface{}    `json:"tags"`
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
