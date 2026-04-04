package services

import (
	"log"
	"net"

	"github.com/pion/stun/v3"
)

type WebRTCService struct {
}

func NewWebRTCService() *WebRTCService {
	return &WebRTCService{}
}

type PacketConnLogger struct {
	net.PacketConn
}

func (p PacketConnLogger) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	if n, err = p.PacketConn.WriteTo(b, addr); err == nil && stun.IsMessage(b) {
		msg := &stun.Message{Raw: b}
		if err = msg.Decode(); err != nil {
			return
		}
		if msg.Type.Class == stun.ClassErrorResponse {
			var code stun.ErrorCodeAttribute
			if err := code.GetFrom(msg); err == nil {
				log.Printf("ERROR %d: %s", code.Code, code.Reason)
			}
		}

		log.Printf("Outbound STUN: %s \n", msg.String())
	}

	return
}

func (p PacketConnLogger) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	if n, addr, err = p.PacketConn.ReadFrom(b); err == nil && stun.IsMessage(b) {
		msg := &stun.Message{Raw: b}
		if err = msg.Decode(); err != nil {
			return
		}

		log.Printf("Inbound STUN: %s \n", msg.String())
	}

	return
}

func (s *WebRTCService) MustCreateUDPListener(ip string) net.PacketConn {
	con, err := net.ListenPacket("udp4", ip)
	if err != nil {
		panic(err)
	}
	return PacketConnLogger{con}
}
