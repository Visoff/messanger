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
	mux.Handle("/conference/room", http.HandlerFunc(c.HandleCreateRoom))

	publicIPs, err := net.LookupIP(os.Getenv("PUBLIC_IP"))
	if err != nil {
		panic(err)
	}
	if len(publicIPs) == 0 {
		panic("PUBLIC_IP is not set")
	}

	turnSecret := os.Getenv("TURN_SECRET")
	if turnSecret == "" {
		turnSecret = "password"
		log.Println("WARNING: TURN_SECRET not set, using default password")
	}

	_, err = turn.NewServer(turn.ServerConfig{
		Realm:              "dev.uni.visoff.ru",
		AllocationLifetime: 5 * time.Minute,
		AuthHandler: func(ra *turn.RequestAttributes) (string, []byte, bool) {
			return ra.Username, turn.GenerateAuthKey(ra.Username, ra.Realm, turnSecret), true
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: webrtc_service.MustCreateUDPListener("0.0.0.0:3478"),
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					Address:      "0.0.0.0",
					RelayAddress: publicIPs[0],
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

func (c *WebRTCController) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	id := uuid.New().String()
	c.roomsMU.Lock()
	c.rooms[id] = &Room{
		id:               id,
		peers:            make(map[string]*Peer),
		peersMU:          sync.RWMutex{},
		trackForwarders:  make(map[string]*TrackForwarder),
		trackForwarderMU: sync.RWMutex{},
	}
	c.roomsMU.Unlock()
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(id))
}

func (c *WebRTCController) HandleRoom(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	if err := validateJWT(token); err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := c.ws_updater.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		log.Println("Room id is not set")
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1008, "room id required"))
		conn.Close()
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

	turnSecret := os.Getenv("TURN_SECRET")
	if turnSecret == "" {
		turnSecret = "password"
	}
	turnUser := fmt.Sprintf("%d", time.Now().Unix())
	turnPass := turn.GenerateAuthKey(turnUser, "dev.uni.visoff.ru", turnSecret)

	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{fmt.Sprintf("stun:%v:3478", os.Getenv("PUBLIC_IP"))}},
			{
				URLs:       []string{fmt.Sprintf("turn:%v:3478", os.Getenv("PUBLIC_IP"))},
				Username:   turnUser,
				Credential: string(turnPass),
			},
		},
		ICETransportPolicy: webrtc.ICETransportPolicyAll,
	}

	pc, err := c.webrtc_service.CreateRTCPeerConnection(peerConnectionConfig)
	if err != nil {
		log.Println(err)
		return
	}

	peer_id := uuid.New().String()
	peer := &Peer{
		id:   peer_id,
		conn: conn,
		pc:   pc,
	}

	done := make(chan struct{})
	defer close(done)

	go func() {
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				peer.connMu.Lock()
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				err := conn.WriteMessage(websocket.PingMessage, nil)
				peer.connMu.Unlock()
				if err != nil {
					return
				}
			}
		}
	}()

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	room.peersMU.Lock()
	room.peers[peer_id] = peer
	room.peersMU.Unlock()

	for k, v := range room.trackForwarders {
		localTrackID := fmt.Sprintf("%s-%s", v.sourcePeerID, v.source.ID())
		localStreamID := fmt.Sprintf("%s-%s", v.sourcePeerID, v.source.StreamID())
		localTrack, err := webrtc.NewTrackLocalStaticRTP(v.source.Codec().RTPCodecCapability, localTrackID, localStreamID)
		if err != nil {
			log.Println(err)
			continue
		}
		_, err = pc.AddTrack(localTrack)
		if err != nil {
			log.Println(err)
			continue
		}
		v.localTracksMU.Lock()
		v.localTracks[k] = localTrack
		v.localTracksMU.Unlock()
	}
	peer.ScheduleRenegotiation()

	defer func() {
		room.peersMU.RLock()
		for _, p := range room.peers {
			if p.id == peer_id {
				continue
			}
			p.connMu.Lock()
			p.conn.WriteJSON(RTCMessage{Type: "peer_left", PeerID: peer_id})
			p.connMu.Unlock()
		}
		room.peersMU.RUnlock()

		peer.connMu.Lock()
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1001, "leaving"))
		peer.connMu.Unlock()
		conn.Close()

		room.trackForwarderMU.Lock()
		for key, fw := range room.trackForwarders {
			if fw.sourcePeerID == peer_id {
				fw.cancel()
				delete(room.trackForwarders, key)
			}
		}
		room.trackForwarderMU.Unlock()

		room.peersMU.Lock()
		delete(room.peers, peer_id)
		room.peersMU.Unlock()

		pc.Close()

		c.roomsMU.Lock()
		room.peersMU.RLock()
		empty := len(room.peers) == 0
		room.peersMU.RUnlock()
		if empty {
			delete(c.rooms, id)
		}
		c.roomsMU.Unlock()
	}()

	pc.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		room.trackForwarderMU.Lock()
		key := fmt.Sprintf("%v:%v:%v", peer_id, tr.PayloadType(), tr.ID())
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
			localTrackID := fmt.Sprintf("%s-%s", forwarder.sourcePeerID, forwarder.source.ID())
			localStreamID := fmt.Sprintf("%s-%s", forwarder.sourcePeerID, forwarder.source.StreamID())
			localTrack, err := webrtc.NewTrackLocalStaticRTP(forwarder.source.Codec().RTPCodecCapability, localTrackID, localStreamID)
			if err != nil {
				log.Println(err)
				continue
			}
			v.pcMu.Lock()
			_, err = v.pc.AddTrack(localTrack)
			v.pcMu.Unlock()
			if err != nil {
				log.Println(err)
				continue
			}
			forwarder.localTracksMU.Lock()
			forwarder.localTracks[k] = localTrack
			forwarder.localTracksMU.Unlock()

			v.ScheduleRenegotiation()
		}

		go func() {
			rtp_buf := make([]byte, 1500)
			for {
				select {
				case <-ctx.Done():
					return
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
							forwarder.localTracksMU.RUnlock()
							return
						}
					}
					forwarder.localTracksMU.RUnlock()
				}
			}
		}()
	})

	pc.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}
		ice := i.ToJSON()

		if ice.SDPMid == nil || *ice.SDPMid == "" {
			log.Println("skipping ICE candidate with empty SDPMid")
			return
		}

		peer.connMu.Lock()
		err := conn.WriteJSON(RTCMessage{Type: "candidate", Candidate: &ice})
		peer.connMu.Unlock()
		if err != nil {
			log.Println("OnICECandidate write error:", err)
		}
	})

	var pendingIceCandidates []*webrtc.ICECandidateInit

	for {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		msg_type, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Println("ws read error:", err)
			}
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
				var pcErr error
				var localDesc *webrtc.SessionDescription
				func() {
					peer.pcMu.Lock()
					defer peer.pcMu.Unlock()
					if err := pc.SetRemoteDescription(*message.Offer); err != nil {
						pcErr = err
						return
					}
					answer, err := pc.CreateAnswer(nil)
					if err != nil {
						pcErr = err
						return
					}
					if err := pc.SetLocalDescription(answer); err != nil {
						pcErr = err
						return
					}
					localDesc = pc.LocalDescription()
				}()
				if pcErr != nil {
					log.Println(pcErr)
					continue
				}

				peer.connMu.Lock()
				err = conn.WriteJSON(RTCMessage{
					Type:   "answer",
					Answer: localDesc,
				})
				peer.connMu.Unlock()
				if err != nil {
					log.Println(err)
					continue
				}

				for _, candidate := range pendingIceCandidates {
					peer.pcMu.Lock()
					if err := pc.AddICECandidate(*candidate); err != nil {
						log.Println(err)
					}
					peer.pcMu.Unlock()
				}
				pendingIceCandidates = nil

			case "answer":
				if message.Answer == nil || message.Answer.SDP == "" {
					log.Println("answer has empty SDP, skipping")
					continue
				}
				var pcErr error
				func() {
					peer.pcMu.Lock()
					defer peer.pcMu.Unlock()
					pcErr = pc.SetRemoteDescription(*message.Answer)
				}()
				if pcErr != nil {
					log.Printf("SetRemoteDescription(answer) error: %v, SDP length: %d", pcErr, len(message.Answer.SDP))
					continue
				}

				for _, candidate := range pendingIceCandidates {
					peer.pcMu.Lock()
					if err := pc.AddICECandidate(*candidate); err != nil {
						log.Println(err)
					}
					peer.pcMu.Unlock()
				}
				pendingIceCandidates = nil

			case "candidate":
				var remoteDesc *webrtc.SessionDescription
				func() {
					peer.pcMu.Lock()
					defer peer.pcMu.Unlock()
					remoteDesc = pc.RemoteDescription()
				}()
				if remoteDesc == nil {
					pendingIceCandidates = append(pendingIceCandidates, message.Candidate)
					continue
				}
				peer.pcMu.Lock()
				if err := pc.AddICECandidate(*message.Candidate); err != nil {
					log.Println(err)
				}
				peer.pcMu.Unlock()

			default:
				log.Println("Unknown message type", message.Type)
			}
		case websocket.PingMessage:
			peer.connMu.Lock()
			conn.WriteMessage(websocket.PongMessage, nil)
			peer.connMu.Unlock()
		case websocket.CloseMessage:
			return
		}
	}
}
