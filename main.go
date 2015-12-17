package main

import (
	"encoding/json"
	"flag"
	"github.com/belogik/goes"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
)

type Item struct {
	Title string  `json:"title"`
	Price float64 `json:"price"`
}

type Results struct {
	Embedded map[string]interface{} `json:"_embedded"`
}

func buildResults(data *goes.Response) *Results {
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
		Embedded: embedded,
	}

	return results
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
		// fmt.Fprintf(w, "Search: %q", q)
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
		}

		extraArgs := make(url.Values, 1)
		searchResults, err := engine.Search(query, []string{"products"}, []string{"product"}, extraArgs)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		responseData := buildResults(searchResults)

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
