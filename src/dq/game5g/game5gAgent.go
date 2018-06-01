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
	"container/list"
	"dq/conf"
	"dq/db"
	"dq/utils"
	"sync"
)

type ShowGame struct {
	GameId int
	Score  int
}

//当前进行的游戏表(展示性)
type ShowGameList struct {
	lock      *sync.RWMutex
	bm        *list.List
	size      int
	limitSize int
}

func NewShowGameList(limitsize int) *ShowGameList {
	return &ShowGameList{
		lock:      new(sync.RWMutex),
		bm:        list.New(),
		size:      0,
		limitSize: limitsize,
	}
}

func (a *ShowGameList) AddShowGame(gameid int, score int) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.bm.Len() == 0 {
		tem := ShowGame{GameId: gameid, Score: score}
		a.bm.PushBack(tem)

		return
	}

	if a.bm.Len() >= a.limitSize {
		little := a.bm.Back()
		littlesg := little.Value.(ShowGame)
		if score <= littlesg.Score {
			return
		}
	}

	for e := a.bm.Front(); e != nil; e = e.Next() {

		sg := e.Value.(ShowGame)
		if score >= sg.Score {
			tem := ShowGame{GameId: gameid, Score: score}
			a.bm.InsertBefore(tem, e)

			if a.bm.Len() > a.limitSize {
				a.bm.Remove(a.bm.Back())
			}

			return
		}

	}

	if a.bm.Len() < a.limitSize {
		tem := ShowGame{GameId: gameid, Score: score}
		a.bm.InsertAfter(tem, a.bm.Back())

	}

}
func (a *ShowGameList) RemoveShowGame(gameid int) {
	a.lock.Lock()
	defer a.lock.Unlock()

	for e := a.bm.Front(); e != nil; e = e.Next() {

		sg := e.Value.(ShowGame)
		if gameid == sg.GameId {

			a.bm.Remove(e)
			return
		}

	}
}
func (a *ShowGameList) GetShowGame(count int) []ShowGame {
	a.lock.RLock()
	defer a.lock.RUnlock()

	showGames := make([]ShowGame, 0)

	addcount := 0

	for e := a.bm.Front(); e != nil; e = e.Next() {

		sg := e.Value.(ShowGame)

		showGames = append(showGames, sg)

		addcount++
		if addcount >= count {
			return showGames
		}

	}
	return showGames
}

//游戏部分
type Game5GAgent struct {
	conn network.Conn

	userdata string

	handles map[string]func(data *datamsg.MsgBase)

	Games    *utils.BeeMap //游戏
	Players  *utils.BeeMap //游戏中的玩家
	Creators *utils.BeeMap //创建了游戏，但是还不是玩家

	Showgames *ShowGameList
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
	a.Creators = utils.NewBeeMap()

	a.Showgames = NewShowGameList(30)

	//time.Time.After()

	a.handles = make(map[string]func(data *datamsg.MsgBase))

	//玩家断线
	a.handles["Disconnect"] = a.DoDisConnectData

	//创建游戏
	a.handles["NewGame"] = a.DoNewGameData

	//自己创建游戏
	a.handles["CS_CreateRoom"] = a.DoCreateRoomData

	//检查是否在游戏中
	a.handles["CheckGame"] = a.DoCheckGameData
	//检查是否在游戏中
	a.handles["CS_CheckGoToGame"] = a.DoCheckGoToGameData

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
	items := a.Showgames.GetShowGame(h2.Count)

	jd := &datamsg.SC_GetGamingInfo{}
	jd.GameInfo = make([]datamsg.MsgGame5GingInfo, 0)
	count := 0
	for _, v := range items {
		if count >= h2.Count {
			break
		}
		//sg := v
		g := a.Games.Get(v.GameId)
		if g == nil {
			continue
		}

		game := g.(*Game5GLogic)
		if game.State == Game5GState_Gaming {
			gameinfo := datamsg.MsgGame5GingInfo{}
			gameinfo.PlayerOneName = game.Player[0].Name
			gameinfo.PlayerTwoName = game.Player[1].Name
			gameinfo.GameId = v.GameId
			gameinfo.Score = v.Score //(game.Player[0].SeasonScore + game.Player[1].SeasonScore) / 2
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

	//检查玩家是否在其他游戏中
	if a.Players.Check(data.Uid) == true {
		player := a.Players.Get(data.Uid)
		if player.(*Game5GPlayer).Game != nil {
			if player.(*Game5GPlayer).Game.GameId != h2.GameId {
				a.WriteMsgBytes(datamsg.NewMsgSC_Result(data.Uid, data.ConnectId, "you have another game!"))
				return
			}
		}
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
	player.SeasonScore = playerinfo.SeasonScore
	player.AvatarUrl = playerinfo.AvatarUrl
	player.firstqiziId = playerinfo.FirstQiZi
	player.secondqiziId = playerinfo.SecondQiZi
	player.qiziId = -1

	db.DbOne.GetPlayerManyInfo(player.Uid, "userbaseinfo", "qizi_move,qizi_move_trail,qizi_floor,qizi_lastplay",
		&player.qizi_move, &player.qizi_move_trail, &player.qizi_floor, &player.qizi_lastplay)

	//	qizi_move       int
	//	qizi_move_trail int
	//	qizi_floor      int
	//	qizi_lastplay   int

	//玩家加入游戏
	if player, err = game.(*Game5GLogic).GoIn(player); err != nil {
		a.WriteMsgBytes(datamsg.NewMsgSC_Result(data.Uid, data.ConnectId, err.Error()))

		return
	}
	a.Players.Set(data.Uid, player)
	a.Creators.Delete(data.Uid)

}

//客户端主动检查是否在游戏中(从其他渠道获取游戏信息,进入游戏)
func (a *Game5GAgent) DoCheckGoToGameData(data *datamsg.MsgBase) {

	h2 := &datamsg.CS_CheckGoToGame{}
	err := json.Unmarshal([]byte(data.JsonData), h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	data.ModeType = "Client"
	data.MsgType = "SC_CheckGoToGame"
	jd := &datamsg.SC_CheckGoToGame{}
	jd.GameId = h2.GameId

	if utils.Milliseconde()-h2.CreateGameTime > 1000*60*30 {
		jd.Code = 0
		jd.Err = "time out"
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))
		return
	}

	game := a.Games.Get(h2.GameId)
	if game == nil {

		jd.Code = 0
		jd.Err = "no game"
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))
		return
	}

	if a.Creators.Check(data.Uid) == false {
		player := a.Players.Get(data.Uid)
		if player == nil {
			jd.Code = 1
			jd.Err = "succ"
			a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))
			return
		}
		jd.Code = 0
		jd.Err = "you have another game"
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))
		return

	} else {
		//		game := a.Games.Get(a.Creators.Get(data.Uid))
		//		if game == nil || game.GameId == h2.GameId{
		//			a.Creators.Delete(data.Uid)
		//			jd.Code = 1
		//			jd.Err = "succ"
		//			a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))
		//			return
		//		}

		jd.Code = 0
		jd.Err = "you have another game"
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))
		return
	}

}

//检查是否在游戏中
func (a *Game5GAgent) DoCheckGameData(data *datamsg.MsgBase) {

	//log.Info("----DoCheckGameData--")
	if a.Creators.Check(data.Uid) == false {
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
	} else {
		game := a.Games.Get(a.Creators.Get(data.Uid))
		if game == nil {
			a.Creators.Delete(data.Uid)
			return
		}
		//game := player.(*Game5GPlayer).Game
		if game.(*Game5GLogic).State >= Game5GState_Result {
			return
		}

		//发送信息
		data1 := &datamsg.MsgBase{}
		data1.ModeType = "Client"
		data1.MsgType = "SC_NewGame"
		data1.Uid = data.Uid
		data1.ConnectId = data.ConnectId
		jd := &datamsg.SC_NewGame{}
		jd.GameId = (a.Creators.Get(data.Uid)).(int)

		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data1, jd))
	}

}
func (a *Game5GAgent) DoCreateRoomData(data *datamsg.MsgBase) {

	if a.Creators.Check(data.Uid) == true {
		return
	}
	a.Creators.Set(data.Uid, -1)

	log.Info("-DoCreateRoomData-data.Uid----%d", data.Uid)

	//time conf.Conf.Game5GInfo.Time
	time := int(conf.Conf.Game5GInfo["CreateRoom_Time"].(float64))
	everytime := int(conf.Conf.Game5GInfo["CreateRoom_EveryTime"].(float64))

	game := NewGame5GLogic_CreateRoom(a, data.Uid, time, everytime, data.Uid)

	a.Games.Set(game.GameId, game)
	a.Creators.Set(data.Uid, game.GameId)

	//
	//发送信息
	data1 := &datamsg.MsgBase{}
	data1.ModeType = "Client"
	data1.MsgType = "SC_NewGame"
	data1.Uid = data.Uid
	data1.ConnectId = data.ConnectId
	jd := &datamsg.SC_NewGame{}
	jd.GameId = game.GameId
	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data1, jd))

}

//
func (a *Game5GAgent) DoNewGameData(data *datamsg.MsgBase) {

	//log.Info("----DoNewGameData--")
	h2 := make(map[string]interface{})
	err := json.Unmarshal([]byte(data.JsonData), &h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	//time conf.Conf.Game5GInfo.Time
	time := int(conf.Conf.Game5GInfo["SeasonMatching_Time"].(float64))
	everytime := int(conf.Conf.Game5GInfo["SeasonMatching_EveryTime"].(float64))

	game := NewGame5GLogic_SeasonMatching(a, int(h2["player1"].(float64)), int(h2["player2"].(float64)), time, everytime)

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
