package lib

import (
	"net/http"
)

func HttpHandler(searcher Searcher, queryBuilder QueryBuilder, presenter Presenter, config map[string]string) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
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
			CdnHost: config["cdn_host"],
		}

		query := queryBuilder.Build(ctx)

		searchResults, err := searcher.Search(query)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonData, err := presenter.Present(searchResults, ctx)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)

	}
}
