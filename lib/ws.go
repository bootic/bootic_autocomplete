package lib

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

type WsError struct {
	Error error `json:"error"`
}

type WsQuery struct {
	Q       string `json:"q"`
	Page    string `json:"page"`
	PerPage string `json:"per_page"`
}

func (q *WsQuery) String() string {
	return fmt.Sprintf("Query: %s", q.Q)
}

func WsHandler(searcher Searcher, queryBuilder QueryBuilder, presenter Presenter, config map[string]string) http.Handler {

	handler := func(ws *websocket.Conn) {

		for {
			var queryMsg WsQuery
			err := websocket.JSON.Receive(ws, &queryMsg)
			if err != nil {
				log.Println("ws closed", err)
				break
			}

			q := queryMsg.Q
			page := PageValue(queryMsg.Page, 1)
			perPage := PageValue(queryMsg.PerPage, 30)

			if perPage > 50 {
				perPage = 50
			}

			ctx := &Context{
				Q:       q,
				Page:    page,
				PerPage: perPage,
				Req:     ws.Request(),
				CdnHost: config["cdn_host"],
			}

			query := queryBuilder.Build(ctx)

			searchResults, err := searcher.Search(query)

			if err != nil {
				websocket.JSON.Send(ws, &WsError{Error: err})
				return
			}

			results := presenter.Present(searchResults, ctx)

			websocket.JSON.Send(ws, results)
		}

		ws.Close()
		// io.Copy(ws, ws)
	}

	return websocket.Handler(handler)
}
