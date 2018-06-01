package hall

import (
	"dq/network"
	"math"

	"dq/datamsg"
	"dq/log"
	"encoding/json"
	"net"
	//"strconv"

	"dq/db"
	"dq/utils"
)

type ScoreAndTime struct {
	Time  int64
	Score int
}

type HallAgent struct {
	conn network.Conn

	userdata string

	handles map[string]func(data *datamsg.MsgBase)

	//
	closeFlag *utils.BeeVar

	//五指棋 玩家匹配表[id] =
	serchPoolFor5G *utils.BeeMap

	ScoreTime []ScoreAndTime
}

type serchInfo struct {
	ConnectId int
	Uid       int

	Time      int //游戏总时间 s
	EveryTime int //游戏每一步棋时间 s

	StartTime int64 //开始匹配时间
	Score     int   //赛季分

	IsAndroid int //是否是机器人

}

func (a *HallAgent) GetConnectId() int {

	return 0
}
func (a *HallAgent) GetModeType() string {
	return ""
}

func (a *HallAgent) Init() {

	a.ScoreTime = make([]ScoreAndTime, 10)
	a.ScoreTime[0] = ScoreAndTime{Time: 0 * 1000, Score: 5}
	a.ScoreTime[1] = ScoreAndTime{Time: 2 * 1000, Score: 20}
	a.ScoreTime[2] = ScoreAndTime{Time: 3 * 1000, Score: 50}
	a.ScoreTime[3] = ScoreAndTime{Time: 5 * 1000, Score: 100}
	a.ScoreTime[4] = ScoreAndTime{Time: 10 * 1000, Score: 200}
	a.ScoreTime[5] = ScoreAndTime{Time: 20 * 1000, Score: 500}
	a.ScoreTime[6] = ScoreAndTime{Time: 30 * 1000, Score: 1000}
	a.ScoreTime[7] = ScoreAndTime{Time: 40 * 1000, Score: 2000}
	a.ScoreTime[8] = ScoreAndTime{Time: 50 * 1000, Score: 4000}
	a.ScoreTime[9] = ScoreAndTime{Time: 60 * 1000, Score: 10000000}

	a.serchPoolFor5G = utils.NewBeeMap()
	a.closeFlag = utils.NewBeeVar(false)

	GetMail().Init()
	GetRank().Init()

	//GetStore().getStoreInfo()Buy
	//GetItemManager().GetItemsInfo(93)

	a.handles = make(map[string]func(data *datamsg.MsgBase))

	a.handles["GetInfo"] = a.DoGetInfoData

	a.handles["CS_GetHallUIInfo"] = a.DoGetHallUIInfoData

	//一场游戏比赛结束
	a.handles["GameOverInfo"] = a.DoGameOverInfoData

	a.handles["CS_GetTskInfo"] = a.DoGetTskInfoData
	a.handles["CS_GetMailInfo"] = a.DoGetMailInfoData
	a.handles["CS_GetStoreInfo"] = a.DoGetStoreInfoData
	a.handles["CS_GetBagInfo"] = a.DoGetBagInfoData
	a.handles["CS_GetRankInfo"] = a.DoGetRankInfoData

	a.handles["CS_GetTaskRewards"] = a.DoGetTaskRewardsData
	a.handles["CS_GetMailRewards"] = a.DoGetMailRewardsData
	a.handles["CS_BuyItem"] = a.DoBuyItemData
	a.handles["CS_ZhuangBeiItem"] = a.DoZhuangBeiData

	a.handles["CS_Share"] = a.DoShareData
	a.handles["CS_Presenter"] = a.DoPresenterData

	a.handles["CS_QuickGame"] = a.DoQuickGameData
	a.handles["CS_QuickGameExit"] = a.DoQuickGameExitData

	//玩家断线
	a.handles["Disconnect"] = a.DoDisConnectData

}

func (a *HallAgent) DoPresenterData(data *datamsg.MsgBase) {
	h2 := &datamsg.CS_Presenter{}
	err := json.Unmarshal([]byte(data.JsonData), h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	if h2.PresenterUid <= 0 {
		return
	}
	//如果是推荐的自己
	if h2.PresenterUid == data.Uid {
		return
	}

	//查看是否已经有推荐者了
	mypre := -1
	db.DbOne.GetPlayerOneOtherInfo(data.Uid, "presenter", &mypre)
	if mypre > 0 {
		return
	}
	db.DbOne.SetPlayerOneOtherInfo(data.Uid, "presenter", h2.PresenterUid)

	initGold := 200
	count := 0
	presenteruid := h2.PresenterUid
	content := "恭喜你,你"

	names := []string{}
	nameone := ""
	db.DbOne.GetPlayerOneOtherInfo(data.Uid, "name", &nameone)
	names = append(names, nameone)
	for {

		nextpresenter := 0
		err = db.DbOne.GetPlayerOneOtherInfo(presenteruid, "presenter", &nextpresenter)
		if err != nil {
			return
		}
		//推荐内容
		contents := content
		for i := len(names) - 1; i >= 0; i-- {
			if i == 0 {
				contents = contents + "推荐了" + names[i]
			} else {
				contents = contents + "推荐的" + names[i]
			}

		}
		//--
		contents = contents + ".请收下你的辛苦费!"
		GetMail().DoPresenterMail(presenteruid, contents, initGold)
		if nextpresenter <= 0 {

			return
		}
		db.DbOne.GetPlayerOneOtherInfo(presenteruid, "name", &nameone)
		names = append(names, nameone)

		presenteruid = nextpresenter
		initGold = int(math.Ceil(float64(float64(initGold) * 0.8)))

		count++
		if count >= 10 {
			return
		}

	}

}

func (a *HallAgent) DoShareData(data *datamsg.MsgBase) {
	GetTaskEveryday().doShare(data.Uid)
}

//DoZhuangBeiData
func (a *HallAgent) DoZhuangBeiData(data *datamsg.MsgBase) {
	h2 := &datamsg.CS_ZhuangBeiItem{}
	err := json.Unmarshal([]byte(data.JsonData), h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	GetItemManager().SetUseItemType(data.Uid, h2.Type)
	//GetStore().Buy(data.Uid, h2.Id,h2.Index)
	//flag := GetStore().Buy(data.Uid, h2.Id, h2.Index)
	flag := true
	if flag == true {
		//更新大厅信息
		playerinfo := &datamsg.MsgPlayerInfo{}
		err := db.DbOne.GetPlayerInfo(data.Uid, playerinfo)
		if err == nil {
			data.ModeType = "Client"
			data.MsgType = "SC_MsgHallInfo"
			jd := datamsg.SC_MsgHallInfo{}
			jd.PlayerInfo = *playerinfo
			a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))
		}

	}

}

//DoBuyItemData
func (a *HallAgent) DoBuyItemData(data *datamsg.MsgBase) {
	h2 := &datamsg.CS_BuyItem{}
	err := json.Unmarshal([]byte(data.JsonData), h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	//GetStore().Buy(data.Uid, h2.Id,h2.Index)
	flag := GetStore().Buy(data.Uid, h2.Id, h2.Index)
	if flag == true {
		//更新大厅信息
		playerinfo := &datamsg.MsgPlayerInfo{}
		err := db.DbOne.GetPlayerInfo(data.Uid, playerinfo)
		if err == nil {
			data.ModeType = "Client"
			data.MsgType = "SC_MsgHallInfo"
			jd := datamsg.SC_MsgHallInfo{}
			jd.PlayerInfo = *playerinfo
			a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))
		}

	}

}

//DoGetMailRewardsData
func (a *HallAgent) DoGetMailRewardsData(data *datamsg.MsgBase) {
	h2 := &datamsg.CS_GetMailRewards{}
	err := json.Unmarshal([]byte(data.JsonData), h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	//GetMail().getMailRewards()
	flag := GetMail().getMailRewards(data.Uid, h2.Id)
	if flag == true {
		//更新大厅信息
		playerinfo := &datamsg.MsgPlayerInfo{}
		err := db.DbOne.GetPlayerInfo(data.Uid, playerinfo)
		if err == nil {
			data.ModeType = "Client"
			data.MsgType = "SC_MsgHallInfo"
			jd := datamsg.SC_MsgHallInfo{}
			jd.PlayerInfo = *playerinfo
			a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))
		}
		//
		data.MsgType = "SC_GetMailRewards"
		jd := datamsg.SC_GetMailRewards{}
		jd.Code = 1
		jd.Id = h2.Id
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))

	}

}

func (a *HallAgent) DoGetTaskRewardsData(data *datamsg.MsgBase) {
	h2 := &datamsg.CS_GetTaskRewards{}
	err := json.Unmarshal([]byte(data.JsonData), h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	flag := GetTaskEveryday().getTskEdRewards(data.Uid, h2.Id)
	if flag == true {
		//更新大厅信息
		playerinfo := &datamsg.MsgPlayerInfo{}
		err := db.DbOne.GetPlayerInfo(data.Uid, playerinfo)
		if err == nil {
			data.ModeType = "Client"
			data.MsgType = "SC_MsgHallInfo"
			jd := datamsg.SC_MsgHallInfo{}
			jd.PlayerInfo = *playerinfo
			a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))
		}
		//
		data.MsgType = "SC_GetTaskRewards"
		jd := datamsg.SC_GetTaskRewards{}
		jd.Code = 1
		jd.Id = h2.Id
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))

	}

}

//
func (a *HallAgent) DoGetRankInfoData(data *datamsg.MsgBase) {

	h2 := &datamsg.CS_GetRankInfo{}
	err := json.Unmarshal([]byte(data.JsonData), h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	data.ModeType = "Client"
	data.MsgType = "SC_RankInfo"

	myscore := 0
	db.DbOne.GetPlayerOneInfo(data.Uid, "userbaseinfo", "seasonscore", &myscore)

	tsdinfo := GetRank().RankInfo(h2.StartRank, h2.EndRank, data.Uid, myscore)
	if tsdinfo != nil {
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, tsdinfo))
	}
}

func (a *HallAgent) DoGetBagInfoData(data *datamsg.MsgBase) {

	data.ModeType = "Client"
	data.MsgType = "SC_BagInfo"

	tsdinfo := GetItemManager().GetItemsInfo(data.Uid)
	if tsdinfo != nil {
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, tsdinfo))
	}
}

func (a *HallAgent) DoGetStoreInfoData(data *datamsg.MsgBase) {

	data.ModeType = "Client"
	data.MsgType = "SC_StoreInfo"
	//GetStore().getStoreInfo()
	tsdinfo := GetStore().getStoreInfo()
	if tsdinfo != nil {
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, tsdinfo))
	}
}

func (a *HallAgent) DoGetMailInfoData(data *datamsg.MsgBase) {

	data.ModeType = "Client"
	data.MsgType = "SC_MailInfo"

	tsdinfo := GetMail().getMailInfo(data.Uid, 20)
	if tsdinfo != nil {
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, tsdinfo))
	}
}

func (a *HallAgent) DoGetTskInfoData(data *datamsg.MsgBase) {

	data.ModeType = "Client"
	data.MsgType = "SC_TskEdInfo"

	tsdinfo := GetTaskEveryday().getTskEdInfo(data.Uid)
	if tsdinfo != nil {
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, tsdinfo))
	}
}

func (a *HallAgent) DoQuickGameData(data *datamsg.MsgBase) {
	//	log.Info("--modeType:"+data.ModeType)
	//	log.Info("--ConnectId:"+strconv.Itoa(data.ConnectId))
	//	log.Info("--MsgId:"+strconv.Itoa(data.MsgId))

	sinfo := &serchInfo{}
	sinfo.ConnectId = data.ConnectId
	sinfo.Uid = data.Uid
	sinfo.Time = 60 * 1  //20分钟
	sinfo.EveryTime = 30 //30秒
	sinfo.StartTime = utils.Milliseconde()
	sinfo.Score = 1000
	sinfo.IsAndroid = 0

	//sinfo.IsAndroid =

	a.serchPoolFor5G.Set(data.Uid, sinfo)

	//回复客户端 搜寻中
	playerinfo := &datamsg.MsgPlayerInfo{}
	err := db.DbOne.GetPlayerInfo(data.Uid, playerinfo)
	if err == nil {
		sinfo.Score = playerinfo.SeasonScore
		sinfo.IsAndroid = playerinfo.IsAndroid

		data.ModeType = "Client"
		data.MsgType = "SC_SerchPlayer"
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, nil))
	}

}

func (a *HallAgent) DoDisConnectData(data *datamsg.MsgBase) {

	log.Info("----DoDisConnectData uid:%d--", data.Uid)
	a.serchPoolFor5G.Delete(data.Uid)

	GetTaskEveryday().DeleteUserTaskEveryday(data.Uid)
	GetItemManager().DeletePlayer(data.Uid)

}

func (a *HallAgent) DoQuickGameExitData(data *datamsg.MsgBase) {

	if a.serchPoolFor5G.Check(data.Uid) {
		a.serchPoolFor5G.Delete(data.Uid)

		//回复客户端 搜寻中
		data.ModeType = "Client"
		data.MsgType = "SC_QuickGameExit"
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, nil))

		return
	}

}

//游戏结束信息
func (a *HallAgent) DoGameOverInfoData(data *datamsg.MsgBase) {

	h2 := &datamsg.GameOverInfo{}
	err := json.Unmarshal([]byte(data.JsonData), h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	//排行数据
	winrank := datamsg.RankNodeInfo{}
	winrank.Uid = h2.WinId
	db.DbOne.GetPlayerManyInfo(h2.WinId, "userbaseinfo", "seasonscore,name,avatarurl", &winrank.Score, &winrank.Name, &winrank.Avatar)
	GetRank().SetValue(winrank)
	loserank := datamsg.RankNodeInfo{}
	loserank.Uid = h2.LoseId
	db.DbOne.GetPlayerManyInfo(h2.LoseId, "userbaseinfo", "seasonscore,name,avatarurl", &loserank.Score, &loserank.Name, &loserank.Avatar)
	GetRank().SetValue(loserank)

	//设置每日任务数据
	GetTaskEveryday().Play(h2.WinId)
	GetTaskEveryday().Play(h2.LoseId)
	GetTaskEveryday().Win(h2.WinId)
	if h2.GameMode == 1 { //好友比赛
		GetTaskEveryday().FriendMatchPlay(h2.WinId)
		GetTaskEveryday().FriendMatchPlay(h2.LoseId)

		GetTaskEveryday().FriendMatchWin(h2.WinId)
	} else if h2.GameMode == 3 { //赛季天梯匹配
		GetTaskEveryday().SeasonMatchPlay(h2.WinId)
		GetTaskEveryday().SeasonMatchPlay(h2.LoseId)
		GetTaskEveryday().SeasonMatchWin(h2.WinId)
	}

	//持久化
	GetTaskEveryday().getPlayer(h2.WinId).writeToDB()
	GetTaskEveryday().getPlayer(h2.LoseId).writeToDB()

	//
	//type GameOverInfo struct {
	//	GameMode   int
	//	WinId      int
	//	LoseId     int
	//	ObserverId []int
	//}
}

//func (a *HallAgent) DoGetHallUIInfoData(data *datamsg.MsgBase) {

//}

func (a *HallAgent) DoGetHallUIInfoData(data *datamsg.MsgBase) {
	//大厅界面信息
	numTed := GetTaskEveryday().getCompleteNumOfTskEd(data.Uid)
	numMail := GetMail().getNewMailNum(data.Uid)
	//log.Info("ted:%d---mail:%d", numTed, numMail)
	if numTed > 0 || numMail > 0 {
		data.ModeType = "Client"
		data.MsgType = "SC_HallUIInfo"
		jd := datamsg.SC_HallUIInfo{}
		jd.TaskED_ShowNum = numTed
		jd.Task_ShowNum = 0
		jd.Mail_ShowNum = numMail
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))
	}
}

func (a *HallAgent) DoGetInfoData(data *datamsg.MsgBase) {

	GetMail().CheckUserPublicMail(data.Uid)
	GetMail().CheckUserMail(data.Uid)

	//ce shi
	//	winrank := datamsg.RankNodeInfo{}
	//	winrank.Uid = data.Uid
	//	db.DbOne.GetPlayerManyInfo(data.Uid, "userbaseinfo", "seasonscore,name,avatarurl", &winrank.Score, &winrank.Name, &winrank.Avatar)
	//	GetRank().SetValue(winrank)

	//a.SendHallUIInfo(data)

	//回复客户端
	playerinfo := &datamsg.MsgPlayerInfo{}
	err := db.DbOne.GetPlayerInfo(data.Uid, playerinfo)
	if err == nil {
		data.ModeType = "Client"
		data.MsgType = "SC_MsgHallInfo"
		jd := datamsg.SC_MsgHallInfo{}
		jd.PlayerInfo = *playerinfo
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))
	} else {
		log.Info(err.Error())
	}

	//检查是否有游戏正在进行中

	a.CheckGame(data.Uid, data.ConnectId)

}

func (a *HallAgent) MatchGetScoreFromTime(time int64) int {
	size := len(a.ScoreTime)
	for i := size - 1; i >= 0; i-- {
		if a.ScoreTime[i].Time < time {
			//log.Info("--getscore:%d", a.ScoreTime[i].Score)
			return a.ScoreTime[i].Score
		}
	}
	return 1000000
}

func (a *HallAgent) Update() {

	//500毫秒循环一次
	oneUpdateTime := 500

	//androidPlayOnce := int64(1000 * 60 * 3)
	androidPlayOnce := int64(1000 * 3)
	//lastAndroidPlayTime := utils.Milliseconde()
	lastAndroidPlayTime := int64(0)

	for {
		t1 := utils.Milliseconde()
		if a.closeFlag.Get() == true {
			return
		}
		serchPlayer := a.serchPoolFor5G.Items()

		//算法忽略
		//loop:
		//
		size := len(serchPlayer)
		if size <= 1 {
			utils.MySleep(t1, int64(oneUpdateTime))
			continue
		}
		//算法开始
		//匹配规则(如果双方匹配总时间超过15秒,)
		fight := [2]int{}
		//i := 0
		for k, v := range serchPlayer {
			//log.Info("--1--serchsize:%d", len(serchPlayer))
			fight[0] = k.(int)
			player1 := v.(*serchInfo)
			maxpipeidu := 1000000            //匹配度(分差最小)
			var pipeiplayer *serchInfo = nil //当前最匹配的玩家var pi *int = nil
			pipeiindex := -1
			delete(serchPlayer, k)
			for k1, v1 := range serchPlayer {
				player2 := v1.(*serchInfo)

				//2个机器人之间需要3分钟
				if player1.IsAndroid == 1 && player2.IsAndroid == 1 {
					if t1-lastAndroidPlayTime < androidPlayOnce {
						continue
					}
				}
				//一个机器人 需要至少5秒
				if player1.IsAndroid == 1 || player2.IsAndroid == 1 {
					if t1-player1.StartTime < 5000 || t1-player2.StartTime < 5000 {
						continue
					}
				}

				alltime := t1 - player1.StartTime + t1 - player2.StartTime
				//log.Info("time %d", alltime)
				scoresub := int(math.Abs(float64(player1.Score - player2.Score)))
				//log.Info("score %d", scoresub)
				//
				//(1000 - scoresub) * (alltime / 20)
				if a.MatchGetScoreFromTime(alltime) > scoresub {

					if maxpipeidu > scoresub {
						maxpipeidu = scoresub
						pipeiplayer = player2
						pipeiindex = k1.(int)
					}
				}
			}
			//匹配成功
			if pipeiplayer != nil {
				delete(serchPlayer, pipeiindex)
				fight[1] = pipeiindex

				//log.Info("--2--serchsize:%d", len(serchPlayer))

				//算法结束
				if a.closeFlag.Get() == true {
					return
				}

				p1 := a.serchPoolFor5G.Get(fight[0])
				p2 := a.serchPoolFor5G.Get(fight[1])
				if p1 != nil && p2 != nil {
					a.serchPoolFor5G.Delete(fight[0])
					a.serchPoolFor5G.Delete(fight[1])
					//机器人之间需要3分钟
					if p1.(*serchInfo).IsAndroid == 1 && p2.(*serchInfo).IsAndroid == 1 {
						lastAndroidPlayTime = utils.Milliseconde()

					}
					log.Info("t1:%d---lastAndroidPlayTime%d---", t1, lastAndroidPlayTime)

					//创建一个游戏
					a.CreateGame(p1.(*serchInfo), p2.(*serchInfo))
					log.Info("CreateGame:%d---%d---", p1.(*serchInfo).Score, p2.(*serchInfo).Score)
				}
			}

			//			i++
			//			if i >= 2 {
			//				break
			//			}
		}

		//时间
		t2 := utils.Milliseconde()
		if t2-t1 >= int64(oneUpdateTime) {
			utils.MySleep(t1, int64(t2-t1+1))
		} else {
			utils.MySleep(t1, int64(oneUpdateTime))
			//goto loop
		}

		//utils.MySleep(t1, int64(oneUpdateTime))

	}
}

func (a *HallAgent) CheckGame(uid int, connectid int) {

	//通知游戏 开始一局新游戏
	data := &datamsg.MsgBase{}
	data.ModeType = "Game5G"
	data.Uid = uid
	data.ConnectId = connectid
	data.MsgType = "CheckGame"

	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, nil))
}

func (a *HallAgent) CreateGame(arg *serchInfo, arg1 *serchInfo) {

	//通知游戏 开始一局新游戏
	data := &datamsg.MsgBase{}
	data.ModeType = "Game5G"
	data.Uid = 0
	data.MsgType = "NewGame"
	jd := make(map[string]interface{})
	jd["player1"] = arg.Uid                 //p1
	jd["player2"] = arg1.Uid                //p2
	jd["player1ConnectId"] = arg.ConnectId  //p1
	jd["player2ConnectId"] = arg1.ConnectId //p2
	jd["time"] = arg.Time
	jd["everytime"] = arg.EveryTime
	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))
}

func (a *HallAgent) Run() {

	a.Init()

	go a.Update()

	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err)
			break
		}

		go a.doMessage(data)

	}
}

func (a *HallAgent) doMessage(data []byte) {
	//log.Info("----------Hall----readmsg---------")
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

func (a *HallAgent) OnClose() {

	GetTaskEveryday().DeleteAll()
	GetItemManager().DeleteAll()
	GetRank().WriteDB()

}

func (a *HallAgent) WriteMsg(msg interface{}) {

}
func (a *HallAgent) WriteMsgBytes(msg []byte) {

	err := a.conn.WriteMsg(msg)
	if err != nil {
		log.Error("write message  error: %v", err)
	}
}
func (a *HallAgent) RegisterToGate() {
	t2 := datamsg.MsgRegisterToGate{
		ModeType: datamsg.HallMode,
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

func (a *HallAgent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *HallAgent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *HallAgent) Close() {
	a.closeFlag.Set(true)
	a.conn.Close()
}

func (a *HallAgent) Destroy() {
	a.conn.Destroy()
}
