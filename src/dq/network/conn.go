package network

import (
	"net"
)

type Conn interface {
	ReadSucc()
	ReadMsg() ([]byte, error)
	WriteMsg(args []byte) error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close()
	Destroy()
	//Write(b []byte)
	//Read(b []byte) (int, error)
}
