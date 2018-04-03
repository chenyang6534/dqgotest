package login
import(
	"dq/network"
	//"dq/gate"
	//"dq/log"
	//"fmt"
	//"dq/datamsg"
	//"errors"
	//"net"
	//"encoding/json"
	//"strconv"
	//"math/rand"
	//"time"
)

type Login struct{
	
	// tcp
	TCPAddr      string
	
	
	TcpClient  *network.TCPClient
}

func (login *Login) Run(closeSig chan bool) {


	
	
	var tcpClient *network.TCPClient
	if login.TCPAddr != "" {
		tcpClient = new(network.TCPClient)
		tcpClient.Addr = login.TCPAddr
		
		tcpClient.NewAgent = func(conn *network.TCPConn) network.Agent {
			a := &LoginAgent{conn: conn}
			a.RegisterToGate()
			return a
		}
	}
	
	if tcpClient != nil {
		login.TcpClient = tcpClient
		tcpClient.Start()
	}
	<-closeSig
	
	if tcpClient != nil {
		tcpClient.Close()
		login.TcpClient = nil
	}
}
