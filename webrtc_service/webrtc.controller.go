package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pion/turn/v5"
	"github.com/pion/webrtc/v4"
)

type Room struct {
	id      string
	peers   map[string]*Peer
	peersMU sync.RWMutex

	trackForwarders  map[string]*TrackForwarder
	trackForwarderMU sync.RWMutex
}

type TrackForwarder struct {
	source       *webrtc.TrackRemote
	sourcePeerID string
	localTracks  map[string]*webrtc.TrackLocalStaticRTP
	localTracksMU sync.RWMutex
	cancel       context.CancelFunc
}

type WebRTCController struct {
	mux            http.Handler
	webrtc_service *WebRTCService
	roomsMU        sync.RWMutex
	rooms          map[string]*Room
	ws_updater     *websocket.Upgrader
}

func NewWebRTCController(ws_updater *websocket.Upgrader, webrtc_service *WebRTCService) *WebRTCController {
	c := &WebRTCController{ws_updater: ws_updater}
	mux := http.NewServeMux()
	c.mux = mux
	c.webrtc_service = webrtc_service
	c.rooms = make(map[string]*Room)

	mux.Handle("/room/{id}", http.HandlerFunc(c.HandleRoom))

	public_id := net.ParseIP(os.Getenv("PUBLIC_IP"))
	if public_id == nil {
		panic("PUBLIC_IP is not set")
	}

	_, err := turn.NewServer(turn.ServerConfig{
		Realm:              "dev.uni.visoff.ru",
		AllocationLifetime: 5 * time.Minute,
		AuthHandler: func(ra *turn.RequestAttributes) (string, []byte, bool) {
			return ra.Username, turn.GenerateAuthKey(ra.Username, ra.Realm, "password"), true
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: webrtc_service.MustCreateUDPListener("0.0.0.0:3478"),
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					Address:      "0.0.0.0",
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

type RTCMessage struct {
	Type      string                     `json:"type"`
	Offer     *webrtc.SessionDescription `json:"offer"`
	Answer    *webrtc.SessionDescription `json:"answer"`
	Candidate *webrtc.ICECandidateInit   `json:"candidate"`
}

func (c *WebRTCController) HandleRoom(w http.ResponseWriter, r *http.Request) {
	conn, err := c.ws_updater.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	id := r.PathValue("id")
	if id == "" {
		log.Println("Room id is not set")
		return
	}

	c.roomsMU.Lock()
	room, ok := c.rooms[id]
	if !ok {
		room = &Room{
			id:               id,
			peers:            make(map[string]*Peer),
			peersMU:          sync.RWMutex{},
			trackForwarders:  make(map[string]*TrackForwarder),
			trackForwarderMU: sync.RWMutex{},
		}
		c.rooms[id] = room
	}
	c.roomsMU.Unlock()

	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{fmt.Sprintf("stun:%v:3478", os.Getenv("PUBLIC_IP"))}},
			{
				URLs:       []string{fmt.Sprintf("turn:%v:3478", os.Getenv("PUBLIC_IP"))},
				Username:   "username_server",
				Credential: "password",
			},
		},
		ICETransportPolicy: webrtc.ICETransportPolicyAll,
	}

	pc, err := c.webrtc_service.CreateRTCPeerConnection(peerConnectionConfig)
	if err != nil {
		log.Println(err)
		return
	}

	peer := &Peer{conn: conn, pc: pc}

	peer_id := uuid.New().String()
	room.peersMU.Lock()
	room.peers[peer_id] = peer
	room.peersMU.Unlock()

	for k, v := range room.trackForwarders {
		localTrack, err := webrtc.NewTrackLocalStaticRTP(v.source.Codec().RTPCodecCapability, v.source.ID(), v.source.StreamID())
		if err != nil {
			log.Println(err)
			continue
		}
		peer.mu.Lock()
		_, err = peer.pc.AddTrack(localTrack)
		peer.mu.Unlock()
		if err != nil {
			log.Println(err)
			continue
		}
		v.localTracksMU.Lock()
		v.localTracks[k] = localTrack
		v.localTracksMU.Unlock()
	}

	defer func() {
		room.peersMU.Lock()
		delete(room.peers, peer_id)
		room.peersMU.Unlock()
		pc.Close()
	}()

	pc.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		room.trackForwarderMU.Lock()
		key := fmt.Sprintf("%v:%v", tr.PayloadType, tr.ID())
		forwarder := &TrackForwarder{
			source:       tr,
			sourcePeerID: peer_id,
			localTracks:  make(map[string]*webrtc.TrackLocalStaticRTP),
		}
		ctx, cancel := context.WithCancel(context.Background())
		forwarder.cancel = cancel
		room.trackForwarders[key] = forwarder
		room.trackForwarderMU.Unlock()

		otherPeers := make(map[string]*Peer)
		room.peersMU.RLock()
		for k, v := range room.peers {
			if k != peer_id {
				otherPeers[k] = v
			}
		}
		room.peersMU.RUnlock()

		for k, v := range otherPeers {
			localTrack, err := webrtc.NewTrackLocalStaticRTP(forwarder.source.Codec().RTPCodecCapability, forwarder.source.ID(), forwarder.source.StreamID())
			if err != nil {
				log.Println(err)
				continue
			}
			v.mu.Lock()
			_, err = v.pc.AddTrack(localTrack)
			v.mu.Unlock()
			if err != nil {
				log.Println(err)
				continue
			}
			forwarder.localTracksMU.Lock()
			forwarder.localTracks[k] = localTrack
			forwarder.localTracksMU.Unlock()

			err = v.Renegotiate()
			if err != nil {
				log.Println(err)
				continue
			}
		}

		go func() {
			rtp_buf := make([]byte, 1500)
			for {
				select {
					case <-ctx.Done():
						return;
					default:
						n, _, err := forwarder.source.Read(rtp_buf)
						if err != nil {
							log.Println(err)
							return
						}
						forwarder.localTracksMU.RLock()
						for _, v := range forwarder.localTracks {
							_, err = v.Write(rtp_buf[:n])
							if err != nil {
								log.Println(err)
								return
							}
						}
						forwarder.localTracksMU.RUnlock()
				}
			}
		}()
	})

	pc.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i != nil {
			peer.mu.Lock()
			ice := i.ToJSON()
			peer.conn.WriteJSON(RTCMessage{Type: "candidate", Candidate: &ice})
			peer.mu.Unlock()
		}
	})

	var pendingIceCandidates []*webrtc.ICECandidateInit

	for {
		msg_type, msg, err := peer.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		switch msg_type {
		case websocket.TextMessage:
			var message RTCMessage
			if err = json.Unmarshal(msg, &message); err != nil {
				log.Println(err)
				continue
			}
			switch message.Type {
			case "offer":
				peer.mu.Lock()
				err := peer.pc.SetRemoteDescription(*message.Offer)
				peer.mu.Unlock()
				if err != nil {
					log.Println(err)
					continue
				}

				peer.mu.Lock()
				answer, err := peer.pc.CreateAnswer(nil)
				peer.mu.Unlock()
				if err != nil {
					log.Println(err)
					continue
				}

				peer.mu.Lock()
				err = peer.pc.SetLocalDescription(answer)
				peer.mu.Unlock()
				if err != nil {
					log.Println(err)
					continue
				}

				peer.mu.Lock()
				err = conn.WriteJSON(RTCMessage{Type: "answer", Answer: &answer})
				peer.mu.Unlock()
				if err != nil {
					log.Println(err)
					continue
				}
				for _, candidate := range pendingIceCandidates {
					peer.mu.Lock()
					err = peer.pc.AddICECandidate(*candidate)
					peer.mu.Unlock()
					if err != nil {
						log.Println(err)
						continue
					}
				}
				pendingIceCandidates = nil
			case "answer":
				peer.mu.Lock()
				err := peer.pc.SetRemoteDescription(*message.Answer)
				peer.mu.Unlock()
				if err != nil {
					log.Println(err)
					continue
				}
			case "candidate":
				peer.mu.Lock()
				offer := peer.pc.RemoteDescription()
				peer.mu.Unlock()
				if offer == nil {
					pendingIceCandidates = append(pendingIceCandidates, message.Candidate)
					continue
				}
				peer.mu.Lock()
				err := peer.pc.AddICECandidate(*message.Candidate)
				peer.mu.Unlock()
				if err != nil {
					log.Println(err)
					continue
				}
			default:
				log.Println("Unknown message type", message.Type)
			}
		case websocket.PingMessage:
			peer.mu.Lock()
			peer.conn.WriteMessage(websocket.PongMessage, nil)
			peer.mu.Unlock()
		case websocket.CloseMessage:
			return
		}
	}
}
