package main

import (
	"flag"
	"fmt"
	"github.com/belogik/goes"
	"github.com/bootic/bootic_autocomplete/lib"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
)

type esQuery struct {
}

func (q *esQuery) Build(ctx *lib.Context) (query map[string]interface{}) {

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

type esSearcher struct {
	conn    *goes.Connection
	esIndex []string
	esType  []string
}

func (s *esSearcher) Search(query map[string]interface{}) (*goes.Response, error) {
	extraArgs := make(url.Values, 1)
	return s.conn.Search(query, s.esIndex, s.esType, extraArgs)
}

func main() {

	maxProcs := runtime.NumCPU()
	runtime.GOMAXPROCS(maxProcs)

	var (
		http_host string
		es_host   string
		cdn_host  string
		es_index  string
		es_type   string
	)

	flag.StringVar(&http_host, "httphost", "localhost:3000", "HTTP host:port for search endpoint")
	flag.StringVar(&es_host, "eshost", "localhost:9200", "HTTP host:port for ElasticSearch server")
	flag.StringVar(&cdn_host, "cdnhost", "https://o.btcdn.co", "CDN host for item images")
	flag.StringVar(&es_index, "esindex", "products", "ElasticSearch index")
	flag.StringVar(&es_type, "estype", "product", "ElasticSearch document type")
	flag.Parse()

	es_host_and_port := strings.Split(es_host, ":")

	config := map[string]string{
		"cdn_host": cdn_host,
	}

	conn := goes.NewConnection(es_host_and_port[0], es_host_and_port[1])

	searcher := &esSearcher{
		conn:    conn,
		esIndex: []string{es_index},
		esType:  []string{es_type},
	}

	presenter := &lib.JsonPresenter{}

	http.Handle("/search", lib.HttpHandler(searcher, &esQuery{}, presenter, config))
	http.Handle("/ws", lib.WsHandler(searcher, &esQuery{}, presenter, config))

	http.Handle("/", http.FileServer(http.Dir("./public")))

	log.Println(fmt.Sprintf("serving http requests on %s", http_host))
	log.Fatal(http.ListenAndServe(http_host, nil))
}
