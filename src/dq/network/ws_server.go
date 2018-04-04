package network

import (
	//"time"
	"dq/log"
	"net/http"
	//"net"
	//"sync"
	//"time"
	//"fmt"
	"dq/utils"

	"github.com/gorilla/websocket"
)

type WSServer struct {
	ServerData
	conns map[*websocket.Conn]struct{}
}

func (server *WSServer) Start() {
	server.init()
	go server.run()
}

func (server *WSServer) GetLoginedConnect() *utils.BeeMap {
	return server.LoginedConnect
}
func (server *WSServer) GetAgents() *utils.BeeMap {
	return server.Agents
}

var upgrader = websocket.Upgrader{
	//ReadBufferSize:    4096,
	//WriteBufferSize:   4096,
	//EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (server *WSServer) SvrConnHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Info("Upgrade err", err)
		return
	}
	log.Info("SvrConnHandler")
	server.mutexConns.Lock()
	if len(server.conns) >= server.MaxConnNum {
		server.mutexConns.Unlock()
		conn.Close()
		log.Debug("too many connections")
		return
	}
	server.conns[conn] = struct{}{}

	server.mutexConns.Unlock()

	server.wgConns.Add(1)

	tcpConn := newWSConn(conn, server.PendingWriteNum, server.msgParser)
	agent := server.NewAgent(tcpConn)

	//server.mutexConns.Lock()
	server.Agents.Set(agent.GetConnectId(), agent)

	//server.mutexConns.Unlock()
	//time.Sleep(10*time.Second)
	//go func() {
	agent.Run()

	// cleanup
	tcpConn.Close()
	server.mutexConns.Lock()
	delete(server.conns, conn)
	//delete(server.Agents, agent.GetCreateId())
	server.mutexConns.Unlock()
	//server.Agents.Delete(agent.GetConnectId())
	agent.OnClose()

	server.wgConns.Done()
	//}()

}

func (server *WSServer) SvrConnHandler1(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("hello"))

}

func (server *WSServer) init() {
	http.HandleFunc("/connect", server.SvrConnHandler)
	http.HandleFunc("/connect1", server.SvrConnHandler1)
	//ln, err := net.Listen("tcp", server.Addr)
	//	if err != nil {
	//		log.Error("%v", err)
	//	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
		log.Debug("invalid MaxConnNum, reset to %v", server.MaxConnNum)
	}
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 100
		log.Debug("invalid PendingWriteNum, reset to %v", server.PendingWriteNum)
	}
	if server.NewAgent == nil {
		log.Error("NewAgent must not be nil")
	}

	//server.ln = ln
	server.conns = make(map[*websocket.Conn]struct{})
	//server.Agents = make(map[int]interface{})
	server.Agents = utils.NewBeeMap()
	server.LoginedConnect = utils.NewBeeMap()

	// msg parser
	msgParser := NewMsgParser()

	server.msgParser = msgParser

	log.Info("------Listen:" + server.Addr)
}

func (server *WSServer) run() {
	server.wgLn.Add(1)
	defer server.wgLn.Done()
	//err := http.ListenAndServeTLS(server.Addr, "bin/conf/214571202380020.pem", "bin/conf/214571202380020.key", nil)
	err := http.ListenAndServe(server.Addr, nil)
	//checkErr(err, "ListenAndServe");
	if err != nil {
		log.Info(err.Error())
	}
	log.Info("Func finish.")
}

func (server *WSServer) Close() {

	//server.ln.Close()
	server.wgLn.Wait()

	server.mutexConns.Lock()
	for conn := range server.conns {
		conn.Close()
	}
	server.conns = nil
	server.mutexConns.Unlock()
	//server.Agents = nil

	server.wgConns.Wait()
	log.Info("tcp Close :%s", server.Addr)
}
