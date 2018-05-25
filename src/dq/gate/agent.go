package gate

import (
	"dq/datamsg"
	"dq/log"
	"dq/network"
	"dq/utils"
	"encoding/json"
	"net"
)

//客户端连接上来的代理

type agent struct {
	conn      network.Conn
	gate      *Gate
	connectId int
	UserData  *utils.BeeMap
}

func (a *agent) GetConnectId() int {

	return a.connectId
}
func (a *agent) GetModeType() string {
	return ""
}
func (a *agent) Run() {
	a.UserData = utils.NewBeeMap()

	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err)
			a.conn.Close()
			break
		}

		h1 := &datamsg.MsgBase{}
		err = json.Unmarshal(data, h1)
		if err != nil {
			log.Info("--error")
			break
		} else {

			a.conn.ReadSucc()
			if h1.MsgType == datamsg.CS_Heart {

				continue
			}
			//log.Info("------readmsg len%d:"+string(data), len(data))
			//设置连接id
			h1.ConnectId = (a.connectId)

			if a.FilterNoLoginMode(h1.ModeType) == false {
				log.Info("you donnot login!!")
				a.conn.Close()
				break
			}
			//设置uid
			if a.UserData.Check("uid") {
				h1.Uid = (a.UserData.Get("uid")).(int)
			}
			//转发数据
			if model := a.gate.Models.Get(h1.ModeType); model != nil {
				//
				//a.ReadDataTime = time.Duration(utils.Milliseconde())
				a.conn.ReadSucc()
				data1, err1 := json.Marshal(h1)
				if err1 == nil {
					model.(*ServersAgent).WriteMsgBytes(data1)
				}

			} else {
				log.Info("not find ModeType:%s", h1.ModeType)
			}

		}

	}
}

//过滤非Login模块的非心跳消息
func (a *agent) FilterNoLoginMode(modetype string) bool {
	if modetype != datamsg.LoginMode {
		if a.UserData.Check("uid") == true {
			return true
		}
		return false
	}

	return true

}

func (a *agent) OnClose() {
	//	if a.gate.AgentChanRPC != nil {
	//		err := a.gate.AgentChanRPC.Call0("CloseAgent", a)
	//		if err != nil {
	//			log.Error("chanrpc error: %v", err)
	//		}
	//	}

	//从登录列表中删除连接id
	if a.UserData.Check("uid") == true {
		connectid := a.gate.TcpServer.GetLoginedConnect().Get(a.UserData.Get("uid"))
		if connectid == nil {
			return
		}
		if connectid == a.connectId {
			a.gate.TcpServer.GetLoginedConnect().Delete(a.UserData.Get("uid"))
			log.Info("--un login--:%v", a.UserData.Get("uid"))
		}

		//给其他模块发送玩家断线消息
		data := &datamsg.MsgBase{}
		data.ConnectId = a.connectId
		data.Uid = a.UserData.Get("uid").(int)
		data.MsgType = "Disconnect"
		for k, v := range a.gate.Models.Items() {
			data.ModeType = k.(string)
			v.(*ServersAgent).WriteMsgBytes(datamsg.NewMsg1Bytes(data, nil))
		}
	}

}

func (a *agent) WriteMsg(msg interface{}) {

}
func (a *agent) WriteMsgBytes(msg []byte) {

	err := a.conn.WriteMsg(msg)
	if err != nil {
		log.Error("write message  error: %v", err)
	}
}

func (a *agent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *agent) Close() {
	a.conn.Close()
}

func (a *agent) Destroy() {
	a.conn.Destroy()
}
