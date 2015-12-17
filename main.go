package main

import (
	"encoding/json"
	"flag"
	"github.com/belogik/goes"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
)

type Item struct {
	Title string  `json:"title"`
	Price float64 `json:"price"`
}

type Results struct {
	TotalItems uint64                 `json:"total_items"`
	Page       uint64                 `json:"page"`
	PerPage    uint64                 `json:"per_page"`
	Embedded   map[string]interface{} `json:"_embedded"`
}

func buildResults(data *goes.Response, meta map[string]uint64) *Results {
	var items []*Item

	dataItems := data.Hits.Hits
	for _, it := range dataItems {
		source := it.Source
		title := source["title"].(string)
		price := source["price"].(float64)

		item := &Item{
			Title: title,
			Price: price,
		}

		items = append(items, item)
	}

	embedded := map[string]interface{}{
		"items": items,
	}

	results := &Results{
		TotalItems: data.Hits.Total,
		Page:       meta["page"],
		PerPage:    meta["per_page"],
		Embedded:   embedded,
	}

	return results
}

func pageValue(rawValue string, defValue uint64) (val uint64) {
	if rawValue != "" {
		pval, err := strconv.ParseUint(rawValue, 10, 64)
		if err == nil {
			val = pval
		} else {
			val = defValue
		}
	} else {
		val = defValue
	}

	return val
}

func main() {

	maxProcs := runtime.NumCPU()
	runtime.GOMAXPROCS(maxProcs)

	var (
		es_host string
	)

	flag.StringVar(&es_host, "eshost", "localhost:9200", "HTTP host:port for ElasticSearch server")
	flag.Parse()

	es_host_and_port := strings.Split(es_host, ":")

	engine := goes.NewConnection(es_host_and_port[0], es_host_and_port[1])

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {

		q := r.URL.Query().Get("q")

		page := pageValue(r.URL.Query().Get("page"), 1)
		perPage := pageValue(r.URL.Query().Get("per_page"), 30)

		size := perPage
		from := perPage * (page - 1)

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"filtered": map[string]interface{}{
					"query": map[string]interface{}{
						"query_string": map[string]interface{}{
							"query":            q,
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
						},
					},
				},
			},
			"size": size,
			"from": from,
		}

		extraArgs := make(url.Values, 1)
		searchResults, err := engine.Search(query, []string{"products"}, []string{"product"}, extraArgs)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		responseData := buildResults(searchResults, map[string]uint64{"page": page, "per_page": perPage})

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
