package main

import (
	"encoding/json"
	"flag"
	"github.com/belogik/goes"
	"github.com/bootic/bootic_autocomplete/lib"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
)

func buildQuery(ctx *lib.Context) (query map[string]interface{}) {

	size := ctx.PerPage
	from := ctx.PerPage * (ctx.Page - 1)

	query = map[string]interface{}{
		"query": map[string]interface{}{
			"filtered": map[string]interface{}{
				"query": map[string]interface{}{
					"query_string": map[string]interface{}{
						"query":            ctx.Q,
						"default_operator": "AND",
					},
				},
				"filter": map[string]interface{}{
					"and": []interface{}{
						map[string]interface{}{
							"term": map[string]interface{}{
								"status": []string{"visible"},
							},
						},
						map[string]interface{}{
							"terms": map[string]interface{}{
								"account.status": []string{"active", "free", "trial"},
							},
						},
					},
				},
			},
		},
		"size": size,
		"from": from,
	}

	return query
}

func main() {

	maxProcs := runtime.NumCPU()
	runtime.GOMAXPROCS(maxProcs)

	var (
		es_host  string
		cdn_host string
	)

	flag.StringVar(&es_host, "eshost", "localhost:9200", "HTTP host:port for ElasticSearch server")
	flag.StringVar(&cdn_host, "cdnhost", "https://o.btcdn.co", "CDN host for item images")
	flag.Parse()

	es_host_and_port := strings.Split(es_host, ":")

	engine := goes.NewConnection(es_host_and_port[0], es_host_and_port[1])

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {

		q := r.URL.Query().Get("q")

		page := lib.PageValue(r.URL.Query().Get("page"), 1)
		perPage := lib.PageValue(r.URL.Query().Get("per_page"), 30)

		if perPage > 50 {
			perPage = 50
		}

		ctx := &lib.Context{
			Q:       q,
			Page:    page,
			PerPage: perPage,
			Req:     r,
			CdnHost: cdn_host,
		}

		query := buildQuery(ctx)

		extraArgs := make(url.Values, 1)
		searchResults, err := engine.Search(query, []string{"products"}, []string{"product"}, extraArgs)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		responseData := lib.BuildResults(searchResults, ctx)

		json_data, err := json.Marshal(responseData)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(json_data)
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
