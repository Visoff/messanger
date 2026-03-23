package controllers

import (
	"net/http"

	"github.com/Visoff/messanger/pkgs/handlers"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WebRTCController struct {
	mux         http.Handler
	ws_updater  *websocket.Upgrader
	rooms       map[uuid.UUID]map[chan []byte]struct{}
}

func NewWebRTCController(ws_updater *websocket.Upgrader) *WebRTCController {
	c := &WebRTCController{ws_updater: ws_updater, rooms: map[uuid.UUID]map[chan []byte]struct{}{}}
	mux := http.NewServeMux()
	c.mux = mux

	mux.Handle("POST /room", handlers.Handler(c.CreateRoom))
	mux.Handle("/room/{id}", handlers.Handler(c.HandleRoom))

	return c
}

func (c *WebRTCController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

func (c *WebRTCController) CreateRoom(w http.ResponseWriter, r *http.Request) error {
	uid, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	c.rooms[uid] = map[chan []byte]struct{}{}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(uid.String()))
	return nil
}

func (c *WebRTCController) HandleRoom(w http.ResponseWriter, r *http.Request) error {
	conn, err := c.ws_updater.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	id, err := handlers.GetParamID(r, "id")
	if err != nil {
		return err
	}

	writer := make(chan []byte)

	if c.rooms[id] == nil {
		c.rooms[id] = map[chan []byte]struct{}{}
	}
	c.rooms[id][writer] = struct{}{}
	defer delete(c.rooms[id], writer)

	for {
		msg_type, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if msg_type == websocket.TextMessage {
			for c := range c.rooms[id] {
				c<- msg
			}
		}
		if msg_type == websocket.CloseMessage {
			break
		}
		if msg_type == websocket.PingMessage {
			conn.WriteMessage(websocket.PongMessage, nil)
		}
	}

	return nil
}
