package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
)

type WebRTCService struct{}

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
	connMu sync.Mutex
	pcMu   sync.Mutex
	conn *websocket.Conn
	pc   *webrtc.PeerConnection

	renegotiateTimer  *time.Timer
}

func (p *Peer) Renegotiate() error {
	log.Println("Renegotiate", p.id)

	p.pcMu.Lock()
	if p.pc.SignalingState() != webrtc.SignalingStateStable {
		state := p.pc.SignalingState()
		p.pcMu.Unlock()
		log.Printf("Renegotiate: PC not stable (state=%v), rescheduling", state)
		p.ScheduleRenegotiation()
		return nil
	}

	offer, err := p.pc.CreateOffer(nil)
	if err != nil {
		p.pcMu.Unlock()
		return err
	}
	err = p.pc.SetLocalDescription(offer)
	if err != nil {
		p.pcMu.Unlock()
		return err
	}

	ld := p.pc.LocalDescription()
	if ld == nil || ld.SDP == "" {
		p.pcMu.Unlock()
		return fmt.Errorf("local description is empty after SetLocalDescription")
	}
	offer = *ld
	p.pcMu.Unlock()

	p.connMu.Lock()
	err = p.conn.WriteJSON(&RTCMessage{
		Type:  "offer",
		Offer: &offer,
	})
	p.connMu.Unlock()
	if err != nil {
		return err
	}

	return nil
}

func (p *Peer) ScheduleRenegotiation() {
	if p.renegotiateTimer != nil {
		p.renegotiateTimer.Stop()
	}
	p.renegotiateTimer = time.AfterFunc(150*time.Millisecond, func() {
		if err := p.Renegotiate(); err != nil {
			log.Println("ScheduleRenegotiation error:", err)
		}
	})
}
