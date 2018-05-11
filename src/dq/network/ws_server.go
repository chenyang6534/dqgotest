package network

import (
	"fmt"
	"time"
	//"time"
	"dq/log"
	"net/http"
	//"net"
	//"sync"
	//"time"
	//"fmt"
	"dq/utils"
	//"reflect"
	"dq/conf"

	"github.com/gorilla/websocket"
)

//type ServerData struct {
//	Addr            string
//	MaxConnNum      int
//	PendingWriteNum int
//	NewAgent        func(Conn) Agent
//	ln              net.Listener

//	//Agents			map[int]interface{}
//	Agents			*utils.BeeMap
//	LoginedConnect	*utils.BeeMap
//	mutexConns      sync.Mutex
//	wgLn            sync.WaitGroup
//	wgConns         sync.WaitGroup

//	// msg parser
//	msgParser    *MsgParser
//}

type WSServer struct {
	ServerData
	//conns        map[*websocket.Conn]struct{}
	conns        *utils.BeeMap
	HttpServer   *http.Server
	isCheckHeart bool
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

	if server.conns.Size() >= server.MaxConnNum {
		conn.Close()
		log.Debug("too many connections")
		return
	}
	conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(conf.Conf.GateInfo.TimeOut)))

	//	server.mutexConns.Lock()
	//	if len(server.conns) >= server.MaxConnNum {
	//		server.mutexConns.Unlock()
	//		conn.Close()
	//		log.Debug("too many connections")
	//		return
	//	}
	//	server.conns[conn] = struct{}{}
	//	server.mutexConns.Unlock()

	server.wgConns.Add(1)

	tcpConn := newWSConn(conn, server.PendingWriteNum, server.msgParser)

	server.conns.Set(tcpConn, struct{}{})

	agent := server.NewAgent(tcpConn)

	//server.mutexConns.Lock()
	server.Agents.Set(agent.GetConnectId(), agent)

	//server.mutexConns.Unlock()
	//time.Sleep(10*time.Second)
	//go func() {
	agent.Run()

	// cleanup
	tcpConn.Close()
	//	server.mutexConns.Lock()
	//	delete(server.conns, conn)
	//	server.mutexConns.Unlock()
	server.conns.Delete(tcpConn)
	agent.OnClose()

	server.wgConns.Done()
	//}()

}

func (server *WSServer) SvrConnHandler1(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("hello"))

}

func (server *WSServer) init() {
	http.HandleFunc("/connect", server.SvrConnHandler)

	server.isCheckHeart = true

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
	//server.conns = make(map[*websocket.Conn]struct{})
	server.conns = utils.NewBeeMap()
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

	//server.wgLn.Add(1)
	//go server.checkHeart()

	server.HttpServer = &http.Server{Addr: server.Addr, Handler: nil}
	{
		//err := http.ListenAndServeTLS(server.Addr, "bin/conf/214571202380020.pem", "bin/conf/214571202380020.key", nil)
		//server.HttpServer := &http.Server{Addr: addr, Handler: handler}
		//err := server.HttpServer.ListenAndServeTLS("bin/conf/214571202380020.pem", "bin/conf/214571202380020.key")
	}

	err := server.HttpServer.ListenAndServe()
	//err := server.HttpServer.ListenAndServeTLS("bin/conf/214571202380020.pem", "bin/conf/214571202380020.key")

	if err != nil {
		log.Info(err.Error())
	}
	fmt.Println("WSServer Func finish.")
}

func (server *WSServer) checkHeart() {

	for {
		if server.isCheckHeart == false {
			fmt.Println("heart over")

			break
		}
		var items = server.conns.Items()
		curtime := time.Duration(utils.Milliseconde())
		for k, _ := range items {
			if server.isCheckHeart == false {
				fmt.Println("heart over")

				break
			}
			//log.Info(reflect.TypeOf(*v).Name())
			//if reflect.TypeOf(*v).Name() == "agent" {
			conn := k.(*WSConn)
			subtime := curtime - conn.ReadDataTime
			//log.Info("subtime:%d", conf.Conf.GateInfo.TimeOut)
			if subtime > time.Duration(conf.Conf.GateInfo.TimeOut*1000) {
				conn.Close()
				log.Info("time out")
			}
			//}

		}

		time.Sleep(time.Second * 3)

	}
	server.wgLn.Done()
}

func (server *WSServer) Close() {

	//fmt.Println("Close11")
	server.isCheckHeart = false

	server.HttpServer.Close()

	server.wgLn.Wait()

	var conns = server.conns.Items()
	for conn := range conns {
		conn.(*WSConn).Close()
	}
	server.conns.DeleteAll()

	//	server.mutexConns.Lock()
	//	for conn := range server.conns {
	//		conn.Close()
	//	}
	//	server.conns = nil
	//	server.mutexConns.Unlock()

	server.wgConns.Wait()
	fmt.Println("tcp Close :%s", server.Addr)
}
