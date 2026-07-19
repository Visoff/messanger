package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/pion/turn/v5"
)

type TURNController struct {
	server *turn.Server
}

func mustCreateUDPListener(addr string) net.PacketConn {
	con, err := net.ListenPacket("udp4", addr)
	if err != nil {
		panic(err)
	}
	return con
}

func NewTURNController() *TURNController {
	publicIPs, err := net.LookupIP(os.Getenv("PUBLIC_IP"))
	if err != nil {
		panic(err)
	}
	if len(publicIPs) == 0 {
		panic("PUBLIC_IP is not set")
	}

	server, err := turn.NewServer(turn.ServerConfig{
		Realm:              "localhost",
		AllocationLifetime: 5 * time.Minute,
		AuthHandler: func(ra *turn.RequestAttributes) (string, []byte, bool) {
			return ra.Username, turn.GenerateAuthKey(ra.Username, ra.Realm, "password"), true
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: mustCreateUDPListener("0.0.0.0:3478"),
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

	return &TURNController{server: server}
}

func (c *TURNController) Close() error {
	return c.server.Close()
}
