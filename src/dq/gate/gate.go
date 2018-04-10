package gate

import (
	"dq/log"
	"dq/network"
	"fmt"

	"dq/utils"
	//"time"
)

type Gate struct {
	MaxConnNum      int
	PendingWriteNum int

	// tcp client
	TCPAddr   string
	WSAddr    string
	TcpServer network.Server
	//TcpServer  	interface{}

	//本地服务器之间通讯
	LocalTCPAddr   string
	LocalTcpServer *network.TCPServer

	//Models	map[string]*ServersAgent
	Models *utils.BeeMap
}

var ConnectId = 10000

//type MessageHandle struct{
//	gate *Gate
//}
//type MessageHandleArgs struct {
//	Session int
//	Msg []byte
//}

//func (t *MessageHandle) MsgHandle(args *MessageHandleArgs, reply *string) error {

//	if t.gate == nil {
//		return errors.New("no find gate")
//	}
//	if t.gate.TcpServer == nil {
//		return errors.New("no find tcpserver")
//	}
//	if _, ok := t.gate.TcpServer.Agents[args.Session]; !ok {
//		return errors.New("no find args.Session")
//	}
//	t.gate.TcpServer.Agents[args.Session].(Agent).WriteMsgBytes(args.Msg)
//    return nil
//}

func (gate *Gate) Run(closeSig chan bool) {

	gate.Models = utils.NewBeeMap()

	var localTcpServer *network.TCPServer
	if gate.LocalTCPAddr != "" {
		localTcpServer = new(network.TCPServer)
		localTcpServer.Addr = gate.LocalTCPAddr
		localTcpServer.MaxConnNum = gate.MaxConnNum
		localTcpServer.PendingWriteNum = gate.PendingWriteNum

		localTcpServer.NewAgent = func(conn network.Conn) network.Agent {
			a := &ServersAgent{conn: conn, gate: gate, ModeType: ""}

			return a
		}

	}
	var wsServer *network.WSServer
	if gate.WSAddr != "" {
		log.Info("WSAddr:" + gate.WSAddr)
		wsServer = new(network.WSServer)
		wsServer.Addr = gate.WSAddr
		wsServer.MaxConnNum = gate.MaxConnNum
		wsServer.PendingWriteNum = gate.PendingWriteNum

		wsServer.NewAgent = func(conn network.Conn) network.Agent {
			//ReadDataTime time.Duration
			a := &agent{conn: conn, gate: gate, connectId: ConnectId}
			ConnectId++

			return a
		}
	}

	var tcpServer *network.TCPServer
	if gate.TCPAddr != "" {
		log.Info("TCPAddr:" + gate.TCPAddr)
		tcpServer = new(network.TCPServer)
		tcpServer.Addr = gate.TCPAddr
		tcpServer.MaxConnNum = gate.MaxConnNum
		tcpServer.PendingWriteNum = gate.PendingWriteNum

		tcpServer.NewAgent = func(conn network.Conn) network.Agent {
			a := &agent{conn: conn, gate: gate, connectId: ConnectId}
			ConnectId++

			return a
		}
	}
	if localTcpServer != nil {
		gate.LocalTcpServer = localTcpServer
		localTcpServer.Start()
	}
	//	var t1 interface{}
	//	t1 = tcpServer
	//	tt := t1.(*network.Server).Addr

	if tcpServer != nil {
		gate.TcpServer = tcpServer
		tcpServer.Start()
	}
	if wsServer != nil {
		gate.TcpServer = wsServer
		wsServer.Start()
	}

	<-closeSig
	if localTcpServer != nil {
		localTcpServer.Close()
		gate.LocalTcpServer = nil
	}

	if tcpServer != nil {
		tcpServer.Close()
		gate.TcpServer = nil
	}
	if wsServer != nil {
		wsServer.Close()

	}
	fmt.Println("gate over")
}

func (gate *Gate) OnDestroy() {}
