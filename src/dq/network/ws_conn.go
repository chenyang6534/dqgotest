package network

import (
	"dq/log"
	"dq/utils"
	"errors"
	"net"
	"sync"
	"time"

	"dq/conf"

	"github.com/gorilla/websocket"
)

//type ConnSet map[net.Conn]struct{}

type WSConn struct {
	sync.Mutex
	conn      *websocket.Conn
	writeChan chan []byte
	closeFlag bool
	msgParser *MsgParser

	ReadDataTime time.Duration
}

func newWSConn(conn *websocket.Conn, pendingWriteNum int, msgParser *MsgParser) *WSConn {
	tcpConn := new(WSConn)
	tcpConn.ReadDataTime = time.Duration(utils.Milliseconde())
	tcpConn.conn = conn
	tcpConn.writeChan = make(chan []byte, pendingWriteNum)
	tcpConn.msgParser = msgParser
	go func() {
		for b := range tcpConn.writeChan {
			if b == nil {
				break
			}
			err := conn.WriteMessage(websocket.TextMessage, b)
			//TextMessage BinaryMessage
			//conn.Write()
			//_, err := conn.Write(b)
			if err != nil {
				break
			}
		}

		conn.Close()
		tcpConn.Lock()
		tcpConn.closeFlag = true
		tcpConn.Unlock()
	}()

	return tcpConn
}

func (tcpConn *WSConn) doDestroy() {
	//tcpConn.conn.(*net.TCPConn).SetLinger(0)
	tcpConn.conn.Close()

	if !tcpConn.closeFlag {
		close(tcpConn.writeChan)
		tcpConn.closeFlag = true
	}
}

func (tcpConn *WSConn) Destroy() {
	tcpConn.Lock()
	defer tcpConn.Unlock()

	tcpConn.doDestroy()
}

func (tcpConn *WSConn) Close() {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag {
		return
	}

	tcpConn.doWrite(nil)
	tcpConn.closeFlag = true
}

func (tcpConn *WSConn) doWrite(b []byte) {
	if len(tcpConn.writeChan) == cap(tcpConn.writeChan) {
		log.Debug("close conn: channel full")
		tcpConn.doDestroy()
		return
	}

	tcpConn.writeChan <- b
}

// b must not be modified by the others goroutines
func (tcpConn *WSConn) Write(b []byte) {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag || b == nil {
		return
	}

	tcpConn.doWrite(b)
}

func (tcpConn *WSConn) Read(b []byte) (int, error) {
	//return tcpConn.conn.Read(b)
	return 0, errors.New("Read error")
}

func (tcpConn *WSConn) LocalAddr() net.Addr {
	return tcpConn.conn.LocalAddr()
}

func (tcpConn *WSConn) RemoteAddr() net.Addr {
	return tcpConn.conn.RemoteAddr()
}

func (tcpConn *WSConn) ReadMsg() ([]byte, error) {
	//return tcpConn.msgParser.Read(tcpConn)

	_, b, err := tcpConn.conn.ReadMessage()

	return b, err
}

func (tcpConn *WSConn) WriteMsg(args []byte) error {
	//return tcpConn.msgParser.Write(tcpConn, args)
	tcpConn.Write(args)
	return nil
}

func (tcpConn *WSConn) ReadSucc() {
	//tcpConn.ReadDataTime = time.Duration(utils.Milliseconde())
	tcpConn.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(conf.Conf.GateInfo.TimeOut)))
}
