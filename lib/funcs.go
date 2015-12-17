package lib

import (
	"fmt"
	"github.com/belogik/goes"
	"github.com/leekchan/accounting"
	"net/http"
	"net/url"
	"strconv"
)

func BuildResults(data *goes.Response, ctx *Context) *Results {
	var items []*Item

	dataItems := data.Hits.Hits
	var itemImage map[string]interface{}

	for _, it := range dataItems {
		source := it.Source
		title := source["title"].(string)
		slug := source["slug"].(string)
		currency := source["currency_code"].(string)
		price := uint64(source["price"].(float64))
		shop := source["shop"].(map[string]interface{})
		shopId := shop["id"].(float64)
		url := shop["url"].(string)

		formattedPrice := formatAmount(price, currency)

		itemLinks := map[string]*Link{
			"web": &Link{Href: fmt.Sprintf("http://%s/products/%s", url, slug)},
		}

		if img, ok := source["image"]; ok {
			itemImage = img.(map[string]interface{})
		}

		itemLinks["thumbnail"] = &Link{Href: itemThumbnail(shopId, ctx.CdnHost, itemImage)}

		item := &Item{
			Links:          itemLinks,
			Title:          title,
			Price:          price,
			FormattedPrice: formattedPrice,
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

var currencies = map[string]accounting.Accounting{
	"CLP": accounting.Accounting{Symbol: "$", Precision: 0, Thousand: ".", Decimal: ","},
}

func formatAmount(amount uint64, currency string) string {
	if ac, ok := currencies[currency]; ok {
		return ac.FormatMoney(amount)
	} else {
		return strconv.FormatUint(amount, 10)
	}
}

func paginationLink(req *http.Request, q string, page, perPage uint64) *Link {
	u := &url.URL{
		Scheme: "http",
		Host:   req.Host,
		Path:   req.URL.Path,
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
	var fileName string
	if fn, ok := image["file_name"]; ok {
		fileName = fn.(string)
	} else {
		return ""
	}
	id := image["id"].(string)

	return fmt.Sprintf("%s/%.0f/small/%s-%s", cdnHost, shopId, id, fileName)
}

func PageValue(rawValue string, defValue uint64) (val uint64) {
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
