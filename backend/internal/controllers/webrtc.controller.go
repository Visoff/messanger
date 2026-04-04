package controllers

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Visoff/messanger/internal/services"
	"github.com/Visoff/messanger/pkgs/handlers"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pion/turn/v5"
)

type WebRTCController struct {
	mux        http.Handler
	ws_updater *websocket.Upgrader
	rooms      map[uuid.UUID]map[chan []byte]struct{}
}

func NewWebRTCController(ws_updater *websocket.Upgrader, webrtc_service *services.WebRTCService) *WebRTCController {
	c := &WebRTCController{ws_updater: ws_updater, rooms: map[uuid.UUID]map[chan []byte]struct{}{}}
	mux := http.NewServeMux()
	c.mux = mux

	mux.Handle("POST /room", handlers.Handler(c.CreateRoom))
	mux.Handle("/room/{id}", http.HandlerFunc(c.HandleRoom))

	public_id := net.ParseIP(os.Getenv("PUBLIC_IP"))
	if public_id == nil {
		panic("PUBLIC_IP is not set")
	}

	_, err := turn.NewServer(turn.ServerConfig{
		Realm: "loc",
		AllocationLifetime: 5 * time.Minute,
		AuthHandler: func(ra *turn.RequestAttributes) (string, []byte, bool) {
			return ra.Username, turn.GenerateAuthKey(ra.Username, ra.Realm, "password"), true
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: webrtc_service.MustCreateUDPListener("0.0.0.0:3478"),
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					Address: "0.0.0.0",
					RelayAddress: public_id,
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	log.Println("TURN server is listening on 0.0.0.0:3478")

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

func (c *WebRTCController) HandleRoom(w http.ResponseWriter, r *http.Request) {
	conn, err := c.ws_updater.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	id, err := handlers.GetParamID(r, "id")
	if err != nil {
		log.Println(err)
		return
	}

	writer := make(chan []byte)
	go func() {
		for msg := range writer {
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}()

	if c.rooms[id] == nil {
		c.rooms[id] = map[chan []byte]struct{}{}
	}
	c.rooms[id][writer] = struct{}{}
	defer delete(c.rooms[id], writer)
	n := len(c.rooms[id])

	for {
		msg_type, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if msg_type == websocket.TextMessage {
			log.Println(n, string(msg))
			for c := range c.rooms[id] {
				if c == writer {
					continue
				}
				c <- msg
			}
		}
		if msg_type == websocket.CloseMessage {
			break
		}
		if msg_type == websocket.PingMessage {
			conn.WriteMessage(websocket.PongMessage, nil)
		}
	}
}
