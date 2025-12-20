package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rustyeddy/otto/messanger"
)

type Websock struct {
	*websocket.Conn
	writeQ chan *messanger.Msg
	Done   chan any
}

var (
	Websocks []*Websock
)

func NewWebsock(conn *websocket.Conn) *Websock {
	ws := &Websock{
		Conn:   conn,
		Done:   make(chan any),
		writeQ: make(chan *messanger.Msg),
	}
	return ws
}

func (ws *Websock) GetWriteQ() chan *messanger.Msg {
	return ws.writeQ
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

func checkOrigin(r *http.Request) bool {
	return true
}

type WServe struct {
}

func (ws WServe) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Info("[I] Connected with Websocket")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Websocket Upgrader err", "error", err)
		return
	}
	defer conn.Close()

	wsock := NewWebsock(conn)
	go func() {
		for {
			// var messanger StationEvent
			mt, message, err := conn.ReadMessage()
			if err != nil {
				slog.Error("websocket read:", "error", err)
				break
			}
			fmt.Printf("%v - %v - %s\n", mt, message, err)
		}
	}()

	Websocks = append(Websocks, wsock)
	wq := wsock.GetWriteQ()
	for {
		select {
		case msg, ok := <-wq:
			fmt.Printf("Recieved message for ws: %+v\n", msg)
			if !ok {
				break
			}

			jbytes, err := msg.JSON()
			if err != nil {
				slog.Error("Failed to JSONify message: ", "error", err)
				continue
			}
			err = conn.WriteJSON(jbytes)
			if err != nil {
				slog.Error("Failed to write web socket", "error", err)
				return
			}

		case <-wsock.Done:
			return
		}
	}
}
