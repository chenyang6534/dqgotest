package login

import (
	"io/ioutil"
)

import (
	"dq/network"
	//"dq/gate"
	"dq/log"
	//"fmt"
	"dq/datamsg"
	//"errors"
	"encoding/json"
	"net"
	//"strconv"
	"dq/db"
	"math/rand"
	"net/http"
	"time"
)

type LoginAgent struct {
	conn network.Conn

	userdata string

	handles map[string]func(data *datamsg.MsgBase)
}

func (a *LoginAgent) registerDataHandle(msgtype string, f func(data *datamsg.MsgBase)) {

	a.handles[msgtype] = f

}

func (a *LoginAgent) GetConnectId() int {

	return 0
}
func (a *LoginAgent) GetModeType() string {
	return ""
}

func (a *LoginAgent) Init() {

	a.handles = make(map[string]func(data *datamsg.MsgBase))
	//a.registerDataHandle("Login",a.DoLoginData)
	a.registerDataHandle("CS_MsgQuickLogin", a.DoQuickLoginData)

	a.registerDataHandle("CS_MsgWeiXingLogin", a.DoWeiXingLoginData)

	rand.Seed(time.Now().UnixNano())
}

func httpGet(appid string, appsecret string, code string) (string, string, error) {
	resp, err := http.Get("https://api.weixin.qq.com/sns/jscode2session?grant_type=authorization_code&appid=" + appid + "&secret=" + appsecret + "&js_code=" + code)
	if err != nil {
		log.Info(err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Info(err.Error())
		return "", "", err
	}
	log.Info(string(body))
	h2 := make(map[string]string)
	err = json.Unmarshal(body, &h2)
	if err != nil {
		log.Info(err.Error())
		return "", "", err
	}
	log.Info(h2["openid"])
	log.Info(h2["session_key"])
	return h2["openid"], h2["session_key"], nil
}

func (a *LoginAgent) DoWeiXingLoginData(data *datamsg.MsgBase) {

	h2 := &datamsg.CS_MsgWeiXingLogin{}
	err := json.Unmarshal([]byte(data.JsonData), h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	openid, _, err1 := httpGet("wx4075bbf91298556f", "ffb292292322a474f81d4095ef2c6edf", h2.Code)
	if err1 != nil {
		log.Info(err1.Error())
		a.WriteMsgBytes(datamsg.NewMsgSC_Result(data.Uid, data.ConnectId, "login faild:"+err1.Error()))
		return
	}
	//return
	//查询数据
	var uid int
	if uid = db.DbOne.CheckWSOpenidLogin(openid); uid > 0 {
		log.Info("---------user login:%d", uid)
	} else {
		log.Info("---------user login name:%s", h2.Name)
		uid = db.DbOne.CreateQuickWSOpenidPlayer(openid, h2.Name)
		if uid < 0 {
			log.Info("---------new user lose", uid)
			return
		}
		log.Info("---------new user:%d", uid)
	}
	//更新头像信息h2.AvatarUrl
	db.DbOne.UpdatePlayerAvatar(h2.AvatarUrl, uid)

	//--------------------
	a.NotifyGateLogined(int(data.ConnectId), uid)

	//db.DbOne.UpdatePlayerWinLose(12,130)

	//回复客户端
	data.ModeType = "Client"
	data.Uid = (uid)
	data.MsgType = "SC_LoginResponse"
	jd := make(map[string]interface{})
	jd["result"] = 1 //成功
	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))

	//通知大厅 新玩家进入
	data.ModeType = "Hall"
	data.Uid = (uid)
	data.MsgType = "GetInfo"
	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, nil))

}

func (a *LoginAgent) DoQuickLoginData(data *datamsg.MsgBase) {
	//	log.Info("--modeType:"+data.ModeType)

	//	log.Info("--ConnectId:"+strconv.Itoa(data.ConnectId))
	//	log.Info("--MsgId:"+strconv.Itoa(data.MsgId))

	h2 := &datamsg.CS_MsgQuickLogin{}
	err := json.Unmarshal([]byte(data.JsonData), h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	//	for i := 0; i < 10; i++{
	//		for j := 0; j < 10; j++{
	//			log.Info("ij:%d",h2.Abc[i][j])

	//		}
	//	}

	//db.dbOne.Query()

	//查询数据
	var uid int
	if uid = db.DbOne.CheckQuickLogin(h2.MachineId, h2.Platform); uid > 0 {
		log.Info("---------user login:%d", uid)
	} else {
		uid = db.DbOne.CreateQuickPlayer(h2.MachineId, h2.Platform, "")
		if uid < 0 {
			log.Info("---------new user lose", uid)
			return
		}
		log.Info("---------new user:%d", uid)
	}
	//--------------------
	a.NotifyGateLogined(int(data.ConnectId), uid)

	//db.DbOne.UpdatePlayerWinLose(12,130)

	//回复客户端
	data.ModeType = "Client"
	data.Uid = (uid)
	data.MsgType = "SC_LoginResponse"
	jd := make(map[string]interface{})
	jd["result"] = 1 //成功
	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))

	//通知大厅 新玩家进入
	data.ModeType = "Hall"
	data.Uid = (uid)
	data.MsgType = "GetInfo"
	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, nil))

}

func (a *LoginAgent) NotifyGateLogined(conncetid int, uid int) {

	data := &datamsg.MsgBase{}
	data.Uid = (uid)
	data.ModeType = datamsg.GateMode
	data.MsgType = "UserLogin"
	data.ConnectId = (conncetid)

	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, nil))

	//	data1,err1 := json.Marshal(data)
	//	if err1 == nil {
	//		a.WriteMsgBytes(data1)
	//	}
}

func (a *LoginAgent) Run() {

	a.Init()

	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err)
			break
		}

		go a.doMessage(data)

	}
}

func (a *LoginAgent) doMessage(data []byte) {
	//log.Info("----------login----readmsg---------")
	h1 := &datamsg.MsgBase{}
	err := json.Unmarshal(data, h1)
	if err != nil {
		log.Info("--error")
	} else {

		//log.Info("--MsgType:" + h1.MsgType)
		if f, ok := a.handles[h1.MsgType]; ok {
			f(h1)
		}

	}

}

func (a *LoginAgent) OnClose() {

}

func (a *LoginAgent) WriteMsg(msg interface{}) {

}
func (a *LoginAgent) WriteMsgBytes(msg []byte) {

	err := a.conn.WriteMsg(msg)
	if err != nil {
		log.Error("write message  error: %v", err)
	}
}
func (a *LoginAgent) RegisterToGate() {
	t2 := datamsg.MsgRegisterToGate{
		ModeType: datamsg.LoginMode,
	}

	t1 := datamsg.MsgBase{
		ModeType: datamsg.GateMode,
		MsgType:  "Register",
	}

	a.WriteMsgBytes(datamsg.NewMsg1Bytes(&t1, &t2))

	//	temp,err := json.Marshal(t2)
	//	t1.JsonData	= string(temp)
	//	data,err := json.Marshal(t1)
	//	if err != nil{
	//		log.Info("-------json.Marshal:"+err.Error())
	//		return
	//	}

	//	a.WriteMsgBytes(data)

}

func (a *LoginAgent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *LoginAgent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *LoginAgent) Close() {
	a.conn.Close()
}

func (a *LoginAgent) Destroy() {
	a.conn.Destroy()
}
