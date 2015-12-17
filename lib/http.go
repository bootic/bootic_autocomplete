package lib

import (
	"github.com/belogik/goes"
	"net/http"
	"net/url"
)

// type Searcher interface {
// 	Search(map[string]interface{}, []string, []string, url.Values) *goes.Results
// }

type QueryBuilder interface {
	Build(*Context) map[string]interface{}
}

func HttpHandler(engine *goes.Conn, buildQuery QueryBuilder) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

	}
}
