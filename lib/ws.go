package lib

import (
	"golang.org/x/net/websocket"
	"io"
)

func WsHandler(ws *websocket.Conn) {
	io.Copy(ws, ws)
}
