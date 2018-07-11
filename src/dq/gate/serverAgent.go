package gate

import (
	"dq/log"
	"dq/network"
	"net"
	//"time"
	//"reflect"
	//"time"
	//"fmt"
	"encoding/json"
	//"dq/rpc"
	//"errors"
	"dq/datamsg"
	//"strconv"
	//"dq/utils"
)

//其他服务器连接上来的代理
type ServersAgent struct {
	conn     network.Conn
	gate     *Gate
	ModeType string
	handles  map[string]func(data *datamsg.MsgBase)

	gateHandles map[string]func(data *datamsg.MsgBase)
}

func (a *ServersAgent) GetConnectId() int {

	return 0
}
func (a *ServersAgent) GetModeType() string {
	return a.ModeType
}

//func (a *ServersAgent) registerDataHandle(msgtype string,f func(data *datamsg.MsgBase)) {

//	a.handles[msgtype] = f

//}

func (a *ServersAgent) Init() {
	a.handles = make(map[string]func(data *datamsg.MsgBase))
	a.handles[datamsg.GateMode] = a.DoGateData
	a.handles[datamsg.ClientMode] = a.DoClientData
	a.handles[datamsg.HallMode] = a.DoHallData
	a.handles[datamsg.Game5GMode] = a.DoGame5GData

	a.gateHandles = make(map[string]func(data *datamsg.MsgBase))
	a.gateHandles["Register"] = a.DoGateRegisterData
	a.gateHandles["UserLogin"] = a.DoGateUserLoginData
}

func (a *ServersAgent) DoGateRegisterData(data *datamsg.MsgBase) {
	h1 := data
	h2 := &datamsg.MsgRegisterToGate{}
	err := json.Unmarshal([]byte(h1.JsonData), h2)
	if err == nil {
		a.ModeType = h2.ModeType
		log.Info("--register modetype:" + h2.ModeType)
		a.gate.Models.Set(h2.ModeType, a)
	}
	//}
}
func (a *ServersAgent) DoGateUserLoginData(data *datamsg.MsgBase) {
	h1 := data
	if h1.Uid <= 0 {
		log.Info("--h1.Uid <= 0--")
		return
	}

	ag1 := (a.gate.TcpServer.GetAgents().Get((h1.ConnectId)))
	if ag1 != nil {
		ag := ag1.(*agent)

		ag.UserData.Set("uid", h1.Uid)
		//检查是否重复登录
		if a.gate.TcpServer.GetLoginedConnect().Check(h1.Uid) {
			//重复登录 删除之前的连接
			connectid := a.gate.TcpServer.GetLoginedConnect().Get(h1.Uid)

			aglast := (a.gate.TcpServer.GetAgents().Get(connectid)).(*agent)
			if aglast != nil {
				//异地登录 强制下线
				log.Info("--异地登录 强制下线--:%v", h1.Uid)
				aglast.Close()
			}
		}
		//设置登录连接表
		a.gate.TcpServer.GetLoginedConnect().Set(h1.Uid, h1.ConnectId)

		log.Info("--login--:%v", h1.Uid)

		return

	}
	log.Info("--not connect %d", (h1.ConnectId))

}

func (a *ServersAgent) DoGateData(data *datamsg.MsgBase) {
	h1 := data
	if f, ok := a.gateHandles[h1.MsgType]; ok {
		f(h1)
	}

}

func (a *ServersAgent) SendToAll(data1 []byte) {

	allAgents := a.gate.TcpServer.GetAgents()
	items := allAgents.Items()
	for _, v := range items {

		ag := v
		if ag != nil {
			ag.(*agent).WriteMsgBytes(data1)

		}
	}
}

func (a *ServersAgent) DoClientData(data *datamsg.MsgBase) {
	h1 := data
	connectid := (h1.ConnectId)
	uid := (h1.Uid)
	//向客户端隐藏connectid 和 uid
	h1.ConnectId = 0
	h1.Uid = 0
	data1, err1 := json.Marshal(h1)
	if err1 == nil {

		//给所有玩家发消息
		if connectid == -2 && uid == -2 {
			go a.SendToAll(data1)
		} else {
			ag := a.gate.TcpServer.GetAgents().Get(connectid)
			if ag == nil {

				items := a.gate.TcpServer.GetLoginedConnect().Items()
				//				for k, v := range items {
				//					log.Info("--uid:%d---connectid:%d---k:%d--v:%d", uid, connectid, k, v.(int))
				//				}
				con := a.gate.TcpServer.GetLoginedConnect().Get(uid)
				if con != nil {
					connectid = (con).(int)
					ag = a.gate.TcpServer.GetAgents().Get(connectid)
				}

			}

			if ag != nil {
				ag.(*agent).WriteMsgBytes(data1)
				//log.Info("send:%s", data.JsonData)
			}
		}

	} else {
		log.Info("--err:%s", err1.Error())
	}
}

func (a *ServersAgent) DoHallData(data *datamsg.MsgBase) {
	h1 := data

	if model := a.gate.Models.Get(h1.ModeType); model != nil {
		data1, err1 := json.Marshal(h1)
		if err1 == nil {
			model.(*ServersAgent).WriteMsgBytes(data1)
		}

	} else {
		log.Info("not find ModeType:%s", h1.ModeType)
	}

}

func (a *ServersAgent) DoGame5GData(data *datamsg.MsgBase) {

	h1 := data

	if model := a.gate.Models.Get(h1.ModeType); model != nil {
		data1, err1 := json.Marshal(h1)
		if err1 == nil {
			model.(*ServersAgent).WriteMsgBytes(data1)
		}

	} else {
		log.Info("not find ModeType:%s", h1.ModeType)
	}

}

func (a *ServersAgent) Run() {
	a.Init()

	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err)
			break
		}
		//log.Info("------readmsg:" + string(data))

		h1 := &datamsg.MsgBase{}
		err = json.Unmarshal(data, h1)
		if err != nil {
			log.Info("--error")
		} else {
			if f, ok := a.handles[h1.ModeType]; ok {
				//go f(h1)
				f(h1)
			}

		}

	}
}

func (a *ServersAgent) OnClose() {

	log.Info("--delete modetype:%s", a.ModeType)
	a.gate.Models.Delete(a.ModeType)
}

func (a *ServersAgent) WriteMsg(msg interface{}) {

}
func (a *ServersAgent) WriteMsgBytes(msg []byte) {

	err := a.conn.WriteMsg(msg)

	if err != nil {
		log.Error("write message  error: %v", err)
	}
}

func (a *ServersAgent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *ServersAgent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *ServersAgent) Close() {
	a.conn.Close()
}

func (a *ServersAgent) Destroy() {
	a.conn.Destroy()
}
