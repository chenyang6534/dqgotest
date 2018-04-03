package dqrpc

import (
	//"errors"
	"fmt"
	"dq/log"
	//"runtime"
	"net/rpc/jsonrpc"
	"net/rpc"
	"net"
)


type RpcServer struct {
	Addr string
	srv *rpc.Server
}


func (server *RpcServer) Start(name string,rcvr interface{}) {
	go server.init(name,rcvr)
}



//func (server *RpcServer) RegisterName(name string,rcvr interface{}) error {
//	if server.srv == nil{
//		return errors.New("server.srv == nil")
//	}
//	if err := server.srv.RegisterName(name, rcvr); err != nil {
//        return err
//    }
//	return nil
//}

func (server *RpcServer) init(name string,rcvr interface{}) error {
	lis, err := net.Listen("tcp", server.Addr)
    if err != nil {
        log.Error("--RpcServer Listen err -:%s",err.Error())
    }
    defer lis.Close()

    srv := rpc.NewServer()
    if err := srv.RegisterName(name, rcvr); err != nil {
        return err
    }
	log.Info("------RpcServer:"+server.Addr)
	server.srv = srv
    for {
        conn, err := lis.Accept()
        if err != nil {
            log.Fatal("lis.Accept(): %v\n", err)
        }
        go server.srv.ServeCodec(jsonrpc.NewServerCodec(conn))
    }

}
