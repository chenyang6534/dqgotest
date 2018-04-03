package hall
import(
	"dq/network"
	
)

type Hall struct{
	
	// tcp
	TCPAddr      string
	
	
	TcpClient  *network.TCPClient
}

func (hall *Hall) Run(closeSig chan bool) {


	
	
	var tcpClient *network.TCPClient
	if hall.TCPAddr != "" {
		tcpClient = new(network.TCPClient)
		tcpClient.Addr = hall.TCPAddr
		
		tcpClient.NewAgent = func(conn *network.TCPConn) network.Agent {
			a := &HallAgent{conn: conn}
			a.RegisterToGate()
			return a
		}
	}
	
	if tcpClient != nil {
		hall.TcpClient = tcpClient
		tcpClient.Start()
	}
	<-closeSig
	
	if tcpClient != nil {
		tcpClient.Close()
		hall.TcpClient = nil
	}
}
