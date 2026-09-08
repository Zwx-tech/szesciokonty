package ws

import (
	"log"
	"net/http"

	"github.com/Zwx-tech/szesciokonty/backend/internal/protocol"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}

func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws upgrade: %v", err)
			return
		}
		defer conn.Close()

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := handle(conn, data); err != nil {
				log.Printf("ws handle: %v", err)
				return
			}
		}
	})
}

func handle(conn *websocket.Conn, data []byte) error {
	if _, err := protocol.DecodeClient(data); err != nil {
		return writeJSON(conn, protocol.NewError("bad_request", err.Error()))
	}
	// Room/match handlers arrive in later tasks.
	return writeJSON(conn, protocol.NewError("not_implemented", "handler not wired yet"))
}

func writeJSON(conn *websocket.Conn, v any) error {
	b, err := protocol.Encode(v)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, b)
}
