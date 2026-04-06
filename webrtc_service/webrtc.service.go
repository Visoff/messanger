package main

import (
	"log"
	"net"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
)

type WebRTCService struct {
}

func NewWebRTCService() *WebRTCService {
	return &WebRTCService{}
}

func (s *WebRTCService) MustCreateUDPListener(ip string) net.PacketConn {
	con, err := net.ListenPacket("udp4", ip)
	if err != nil {
		panic(err)
	}
	return con
}

func (s *WebRTCService) CreateRTCPeerConnection(config webrtc.Configuration) (*webrtc.PeerConnection, error) {
	mediaEngine := &webrtc.MediaEngine{}
	if err := mediaEngine.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}

	interceptorRegistry := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		return nil, err
	}

	intervalPliFactory, err := intervalpli.NewReceiverInterceptor()
	if err != nil {
		return nil, err
	}
	interceptorRegistry.Add(intervalPliFactory)

	peer, err := webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(interceptorRegistry),
	).NewPeerConnection(config)

	return peer, err
}

type Peer struct {
	id   string
	mu   sync.RWMutex
	conn *websocket.Conn
	pc   *webrtc.PeerConnection

	pendingCandidates []webrtc.ICECandidate
}

func (p *Peer) Renegotiate() error {
	log.Println("Renegotiate", p.id)
	p.mu.Lock()
	defer p.mu.Unlock()
	offer, err := p.pc.CreateOffer(nil)
	if err != nil {
		return err
	}

	err = p.pc.SetLocalDescription(offer)
	if err != nil {
		return err
	}

	p.pendingCandidates = nil

	err = p.conn.WriteJSON(&RTCMessage{
		Type: "offer",
		Offer: p.pc.LocalDescription(),
	})
	if err != nil {
		return err
	}

	return nil
}
