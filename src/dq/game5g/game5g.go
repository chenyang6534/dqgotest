package game5g

import (
	"fmt"
)
import (
	"dq/network"
)

type Game5G struct {

	// tcp
	TCPAddr string

	TcpClient *network.TCPClient
}

func (game5g *Game5G) Run(closeSig chan bool) {

	var tcpClient *network.TCPClient
	if game5g.TCPAddr != "" {
		tcpClient = new(network.TCPClient)
		tcpClient.Addr = game5g.TCPAddr

		tcpClient.NewAgent = func(conn *network.TCPConn) network.Agent {
			a := &Game5GAgent{conn: conn}
			a.RegisterToGate()
			return a
		}
	}

	if tcpClient != nil {
		game5g.TcpClient = tcpClient
		tcpClient.Start()
	}
	<-closeSig

	if tcpClient != nil {
		tcpClient.Close()
		game5g.TcpClient = nil
	}
	fmt.Println("game5g over")
}
