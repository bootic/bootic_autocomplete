package main

import (
	"encoding/json"
	"flag"
	"github.com/belogik/goes"
	"log"
	"net/http"
	"runtime"
	"strings"
)

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

		searchResults, err := engine.Search(query, []string{"products"}, []string{"product"})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json_data, err := json.Marshal(searchResults)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(json_data)
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
