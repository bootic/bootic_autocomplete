package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/belogik/goes"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
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

func buildResults(data *goes.Response, ctx *Context) *Results {
	var items []*Item

	dataItems := data.Hits.Hits
	var itemImage map[string]interface{}

	for _, it := range dataItems {
		source := it.Source
		title := source["title"].(string)
		slug := source["slug"].(string)
		price := source["price"].(float64)
		shop := source["shop"].(map[string]interface{})
		shopId := shop["id"].(float64)
		url := shop["url"].(string)
		itemLinks := map[string]*Link{
			"btc:web": &Link{Href: fmt.Sprintf("http://%s/products/%s", url, slug)},
		}

		if img, ok := source["image"]; ok {
			itemImage = img.(map[string]interface{})
		}

		itemLinks["btc:thumbnail"] = &Link{Href: itemThumbnail(shopId, ctx.CdnHost, itemImage)}

		item := &Item{
			Links: itemLinks,
			Title: title,
			Price: price,
		}

		items = append(items, item)
	}

	var embeddedItems interface{}

	if len(items) == 0 {
		embeddedItems = []string{}
	} else {
		embeddedItems = items
	}

	page := ctx.Page
	perPage := ctx.PerPage
	q := ctx.Q

	embedded := map[string]interface{}{
		"items": embeddedItems,
	}

	var nextPage uint64

	tot := (page-1)*perPage + uint64(len(items))
	if tot < data.Hits.Total {
		nextPage = page + 1
	}

	links := map[string]interface{}{}

	links["self"] = paginationLink(ctx.Req, q, page, perPage)

	if nextPage > 0 {
		links["next"] = paginationLink(ctx.Req, q, nextPage, perPage)
	}

	if page > 1 {
		links["prev"] = paginationLink(ctx.Req, q, page-1, perPage)
	}

	results := &Results{
		Links:      links,
		TotalItems: data.Hits.Total,
		Page:       page,
		PerPage:    perPage,
		Embedded:   embedded,
	}

	return results
}

func paginationLink(req *http.Request, q string, page, perPage uint64) *Link {
	u := &url.URL{
		Scheme: "http",
		Host:   req.Host,
		Path:   "search",
	}

	query := u.Query()
	query.Set("page", strconv.FormatUint(page, 10))
	query.Set("per_page", strconv.FormatUint(perPage, 10))
	query.Set("q", q)
	u.RawQuery = query.Encode()

	return &Link{Href: u.String()}
}

func itemThumbnail(shopId float64, cdnHost string, image map[string]interface{}) string {
	// https://o.btcdn.co/224/small/25368-stallion2.gif
	fileName := image["file_name"].(string)
	id := image["id"].(string)

	return fmt.Sprintf("https://%s/%.0f/small/%s-%s", cdnHost, shopId, id, fileName)
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

		page := pageValue(r.URL.Query().Get("page"), 1)
		perPage := pageValue(r.URL.Query().Get("per_page"), 30)

		if perPage > 50 {
			perPage = 50
		}

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

		extraArgs := make(url.Values, 1)
		searchResults, err := engine.Search(query, []string{"products"}, []string{"product"}, extraArgs)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ctx := &Context{
			Q:       q,
			Page:    page,
			PerPage: perPage,
			Req:     r,
			CdnHost: cdn_host,
		}

		responseData := buildResults(searchResults, ctx)

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
