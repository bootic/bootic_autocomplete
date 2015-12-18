package lib

import (
	"encoding/json"
	"net/http"
)

type httpHandler struct {
	searcher     Searcher
	queryBuilder QueryBuilder
	presenter    Presenter
	config       map[string]string
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query().Get("q")

	page := PageValue(r.URL.Query().Get("page"), 1)
	perPage := PageValue(r.URL.Query().Get("per_page"), 30)

	if perPage > 50 {
		perPage = 50
	}

	ctx := &Context{
		Q:       q,
		Page:    page,
		PerPage: perPage,
		Req:     r,
		CdnHost: h.config["cdn_host"],
	}

	query := h.queryBuilder.Build(ctx)

	searchResults, err := h.searcher.Search(query)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	results := h.presenter.Present(searchResults, ctx)

	jsonData, err := json.Marshal(results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func HttpHandler(searcher Searcher, queryBuilder QueryBuilder, presenter Presenter, config map[string]string) http.Handler {

	return &httpHandler{
		searcher:     searcher,
		queryBuilder: queryBuilder,
		presenter:    presenter,
		config:       config,
	}

}
