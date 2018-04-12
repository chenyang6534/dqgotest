package game5g

import (
	"dq/network"
	"strconv"

	"dq/datamsg"
	"dq/log"
	"encoding/json"
	"net"
	//"strconv"
	//"time"
	"dq/db"
	"dq/utils"
)

type Game5GAgent struct {
	conn network.Conn

	userdata string

	handles map[string]func(data *datamsg.MsgBase)

	Games   *utils.BeeMap
	Players *utils.BeeMap
}

func (a *Game5GAgent) GetConnectId() int {

	return 0
}
func (a *Game5GAgent) GetModeType() string {
	return ""
}

func (a *Game5GAgent) Init() {

	a.Games = utils.NewBeeMap()
	a.Players = utils.NewBeeMap()

	//time.Time.After()

	a.handles = make(map[string]func(data *datamsg.MsgBase))

	//玩家断线
	a.handles["Disconnect"] = a.DoDisConnectData

	//创建游戏
	a.handles["NewGame"] = a.DoNewGameData

	//检查是否在游戏中
	a.handles["CheckGame"] = a.DoCheckGameData

	//玩家进来
	a.handles["CS_GoIn"] = a.DoGoInData
	//玩家退出游戏
	a.handles["CS_GoOut"] = a.DoGoOutData

	//玩家走棋
	a.handles["CS_DoGame5G"] = a.DoDoGame5GData

	//获取当前正在进行的游戏信息
	a.handles["CS_GetGamingInfo"] = a.DoGetGamingInfoData

}

func (a *Game5GAgent) DoGetGamingInfoData(data *datamsg.MsgBase) {

	h2 := &datamsg.CS_GetGamingInfo{}
	err := json.Unmarshal([]byte(data.JsonData), h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	//---------------
	//if( a.Games > 0)
	items := a.Games.Items()

	jd := &datamsg.SC_GetGamingInfo{}
	jd.GameInfo = make([]datamsg.MsgGame5GingInfo, 0)
	count := 0
	for k, v := range items {
		if count >= h2.Count {
			break
		}
		game := v.(*Game5GLogic)
		if game.State == Game5GState_Gaming {
			gameinfo := datamsg.MsgGame5GingInfo{}
			gameinfo.PlayerOneName = game.Player[0].Name
			gameinfo.PlayerTwoName = game.Player[1].Name
			gameinfo.GameId = k.(int)
			gameinfo.Score = 1000
			gameinfo.GameName = "game_" + strconv.Itoa(gameinfo.GameId)
			jd.GameInfo = append(jd.GameInfo, gameinfo)
			count++
		}

	}
	data.ModeType = "Client"
	data.MsgType = "SC_GetGamingInfo"

	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))

}

////获取当前进行中的游戏信息
//type CS_GetGamingInfo struct {
//	Count int //数量
//}

////当前进行中的游戏信息
//type MsgGame5GingInfo struct {
//	GameId        int
//	PlayerOneName string
//	PlayerTwoName string
//	Score         int
//}

////当前进行中的游戏信息
//type SC_GetGamingInfo struct {
//	GameInfo []MsgGame5GingInfo
//}

func (a *Game5GAgent) DoDoGame5GData(data *datamsg.MsgBase) {

	h2 := &datamsg.CS_DoGame5G{}
	err := json.Unmarshal([]byte(data.JsonData), h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	//---------------

	player := a.Players.Get(data.Uid)
	if player == nil {
		a.WriteMsgBytes(datamsg.NewMsgSC_Result(data.Uid, data.ConnectId, "no player"))
		return
	}

	game := player.(*Game5GPlayer).Game
	if game == nil {
		a.WriteMsgBytes(datamsg.NewMsgSC_Result(data.Uid, data.ConnectId, "no game!"))

		return
	}

	//玩家走棋
	if err = game.DoGame5G(player.(*Game5GPlayer).SeatIndex, h2); err != nil {
		a.WriteMsgBytes(datamsg.NewMsgSC_Result(data.Uid, data.ConnectId, "cannot go in game"))

		return
	}

}

func (a *Game5GAgent) DoGoOutData(data *datamsg.MsgBase) {

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	if player.(*Game5GPlayer).Game == nil {
		return
	}
	game := player.(*Game5GPlayer).Game
	if game.State >= Game5GState_Result {
		return
	}

	//玩家退出游戏
	if ok := game.GoOut(player.(*Game5GPlayer)); ok {
		a.Players.Delete(data.Uid)
		return
	}

}

func (a *Game5GAgent) DoGoInData(data *datamsg.MsgBase) {

	h2 := &datamsg.CS_GoIn{}
	err := json.Unmarshal([]byte(data.JsonData), h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	//---------------
	game := a.Games.Get(h2.GameId)
	if game == nil {
		a.WriteMsgBytes(datamsg.NewMsgSC_Result(data.Uid, data.ConnectId, "no game!"))

		return
	}
	if game.(*Game5GLogic).State >= Game5GState_Result {
		a.WriteMsgBytes(datamsg.NewMsgSC_Result(data.Uid, data.ConnectId, "game over!"))

		return
	}

	//创建玩家
	playerinfo := &datamsg.MsgPlayerInfo{}
	err1 := db.DbOne.GetPlayerInfo(data.Uid, playerinfo)
	if err1 != nil {
		log.Info(err1.Error())
		return
	}
	player := &Game5GPlayer{}
	player.Uid = data.Uid
	player.ConnectId = data.ConnectId
	player.Gold = playerinfo.Gold
	player.LoseCount = playerinfo.LoseCount
	player.Name = playerinfo.Name
	player.WinCount = playerinfo.WinCount

	//玩家加入游戏
	if player, err = game.(*Game5GLogic).GoIn(player); err != nil {
		a.WriteMsgBytes(datamsg.NewMsgSC_Result(data.Uid, data.ConnectId, err.Error()))

		return
	}
	a.Players.Set(data.Uid, player)

}

//检查是否在游戏中
func (a *Game5GAgent) DoCheckGameData(data *datamsg.MsgBase) {

	//log.Info("----DoCheckGameData--")
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	if player.(*Game5GPlayer).Game == nil {
		return
	}
	game := player.(*Game5GPlayer).Game
	if game.State >= Game5GState_Result {
		return
	}

	//发送信息
	data1 := &datamsg.MsgBase{}
	data1.ModeType = "Client"
	data1.MsgType = "SC_NewGame"
	data1.Uid = data.Uid
	data1.ConnectId = data.ConnectId
	jd := &datamsg.SC_NewGame{}
	jd.GameId = player.(*Game5GPlayer).Game.GameId

	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data1, jd))

}

//CheckGame
func (a *Game5GAgent) DoNewGameData(data *datamsg.MsgBase) {

	log.Info("----DoNewGameData--")
	//	jd["player1"] = arg.Uid //p1
	//	jd["player2"] = arg1.Uid //p2
	//	jd["player1ConnectId"] = arg.ConnectId //p1
	//	jd["player2ConnectId"] = arg1.ConnectId //p2
	//	jd["time"] = arg.Time
	//	jd["everytime"] = arg.EveryTime

	h2 := make(map[string]interface{})
	err := json.Unmarshal([]byte(data.JsonData), &h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	game := NewGame5GLogic(a, int(h2["player1"].(float64)), int(h2["player2"].(float64)), int(h2["time"].(float64)), int(h2["everytime"].(float64)))

	a.Games.Set(game.GameId, game)

	//
	//发送信息
	data1 := &datamsg.MsgBase{}
	data1.ModeType = "Client"
	data1.MsgType = "SC_NewGame"
	data1.Uid = int(h2["player1"].(float64))
	data1.ConnectId = int(h2["player1ConnectId"].(float64))
	jd := &datamsg.SC_NewGame{}
	jd.GameId = game.GameId
	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data1, jd))

	data1.Uid = int(h2["player2"].(float64))
	data1.ConnectId = int(h2["player2ConnectId"].(float64))
	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data1, jd))
}

func (a *Game5GAgent) DoDisConnectData(data *datamsg.MsgBase) {

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	if player.(*Game5GPlayer).Game == nil {
		return
	}
	game := player.(*Game5GPlayer).Game
	if game.State >= Game5GState_Result {
		return
	}

	//玩家退出游戏
	if ok := game.Disconnect(player.(*Game5GPlayer)); ok {
		//a.Players.Delete(data.Uid)
		return
	}

}

func (a *Game5GAgent) Run() {

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

func (a *Game5GAgent) doMessage(data []byte) {
	//log.Info("----------game5g----readmsg---------")
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

func (a *Game5GAgent) OnClose() {

}

func (a *Game5GAgent) WriteMsg(msg interface{}) {

}
func (a *Game5GAgent) WriteMsgBytes(msg []byte) {

	err := a.conn.WriteMsg(msg)
	if err != nil {
		log.Error("write message  error: %v", err)
	}
}
func (a *Game5GAgent) RegisterToGate() {
	t2 := datamsg.MsgRegisterToGate{
		ModeType: datamsg.Game5GMode,
	}

	t1 := datamsg.MsgBase{
		ModeType: datamsg.GateMode,
		MsgType:  "Register",
	}

	a.WriteMsgBytes(datamsg.NewMsg1Bytes(&t1, &t2))

}

func (a *Game5GAgent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *Game5GAgent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *Game5GAgent) Close() {
	a.conn.Close()
}

func (a *Game5GAgent) Destroy() {
	a.conn.Destroy()
}
