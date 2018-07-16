package hall

import (
	"dq/conf"
	"dq/datamsg"
	"dq/db"
	"dq/log"
	"dq/utils"
	"math/rand"
	"sync"
	"time"
)

var (
	turnTableLogic = &TurnTableLogic{Lock: new(sync.RWMutex), IdValue: utils.NewBeeMap()}
)

type TurnTableLogic struct {
	Lock *sync.RWMutex

	IdValue *utils.BeeMap
}

func GetTurnTableLogic() *TurnTableLogic {
	return turnTableLogic
}

//获取商店信息
func (turntab *TurnTableLogic) GetTurnTableInfo(uid int) *datamsg.SC_TurnTableInfo {
	log.Info("TurnTableLogic")
	turntab.Lock.RLock()
	defer turntab.Lock.RUnlock()

	tt := conf.GetTurnTableConfig()
	for k, v := range tt.TurnTables {
		//log.Info("v----%v", v)
		if turntab.IdValue.Get(v.Id) == nil {
			turntab.IdValue.Set(v.Id, v.InitLuckValue)
		}
		tt.TurnTables[k].Value = (turntab.IdValue.Get(v.Id)).(int)
	}
	//看视频免费抽
	lookviewtime := "2006-01-02"
	lookviewcount := 0
	db.DbOne.GetPlayerManyInfo(uid, "userbaseinfo", "lookviewtime,lookviewcount",
		&lookviewtime, &lookviewcount)
	today := time.Now().Format("2006-01-02")
	if today != lookviewtime {
		lookviewcount = 0
		lookviewtime = today
		db.DbOne.SetPlayerOneInfo(uid, "userbaseinfo", "lookviewcount", lookviewcount)
		db.DbOne.SetPlayerOneInfo(uid, "userbaseinfo", "lookviewtime", lookviewtime)

	}
	LookViewRemainder := conf.GetNoticeConfig().LookViewCount - lookviewcount

	//时间到免费抽
	lastturntabletime := "2006-01-02 15:04:05"
	db.DbOne.GetPlayerOneInfo(uid, "userbaseinfo", "lastturntabletime", &lastturntabletime)

	lasttimeT, _ := time.Parse("2006-01-02 15:04:05", lastturntabletime)
	lasttime := lasttimeT.Unix()

	nowtimeT, _ := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"))

	nowtime := nowtimeT.Unix()
	subtime := nowtime - lasttime

	log.Info("nowtime:%d---lasttime:%d-- subtime:%d", nowtime, lasttime, subtime)

	//时间
	if subtime >= 60*60*24 {
		jd := &datamsg.SC_TurnTableInfo{tt, 0, LookViewRemainder}
		return jd
	} else {
		jd := &datamsg.SC_TurnTableInfo{tt, 60*60*24 - int(subtime), LookViewRemainder}
		return jd
	}

}

//获取商店信息
func (turntab *TurnTableLogic) GetFreeTurnTable(uid int) int {

	lastturntabletime := "2006-01-02 15:04:05"
	db.DbOne.GetPlayerOneInfo(uid, "userbaseinfo", "lastturntabletime", &lastturntabletime)

	lasttimeT, _ := time.Parse("2006-01-02 15:04:05", lastturntabletime)
	lasttime := lasttimeT.Unix()

	nowtimeT, _ := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"))

	nowtime := nowtimeT.Unix()
	subtime := nowtime - lasttime

	//时间
	if subtime >= 60*60*24 {
		return 1
	} else {
		return 0
	}

}

func (turntab *TurnTableLogic) getRandTurnTable(tt []conf.TurnTable) conf.TurnTable {
	//总幸运值
	allnum := 0
	for _, v := range tt {
		allnum += v.Value
	}
	//随机
	randnum := rand.Intn(allnum)
	donum := 0
	for _, v := range tt {
		donum += v.Value
		if randnum < donum {

			return v
		}
	}
	return tt[len(tt)-1]
}

//更改概率
func (turntab *TurnTableLogic) changeValue(tt []conf.TurnTable, ttone conf.TurnTable) {

	allAdd := 0

	ttlevel3 := make([]conf.TurnTable, 0)

	//增加概率
	for _, v := range tt {
		if v.Id != ttone.Id {
			allAdd += v.Add

			turntab.IdValue.AddInt(v.Id, v.Add)
		} else {

		}
		if v.Level == 3 {
			ttlevel3 = append(ttlevel3, v)
		}
	}

	//降低概率
	idValue := turntab.IdValue.Get(ttone.Id)
	if idValue.(int) > allAdd {
		turntab.IdValue.AddInt(ttone.Id, 0-allAdd)
	} else {
		turntab.IdValue.AddInt(ttone.Id, 0)
	}

	//重新初始化 幸运值

	if ttone.Level == 1 || ttone.Level == 2 {
		curvalue := turntab.IdValue.Get(ttone.Id).(int)
		subval := curvalue - ttone.InitLuckValue
		turntab.IdValue.Set(ttone.Id, ttone.InitLuckValue)
		for _, v := range ttlevel3 {
			turntab.IdValue.AddInt(v.Id, int(subval/len(ttlevel3)))
		}
	}

}

////通知所有玩家中奖消息
func (turntab *TurnTableLogic) NoticeAll(uid int, ttone conf.TurnTable, a *HallAgent) {

	if ttone.Level == 1 {
		//通知所有玩家中奖消息
		playername := "player"
		db.DbOne.GetPlayerOneInfo(uid, "userbaseinfo", "name", &playername)
		jdsw := &datamsg.SC_ScrollWords{}
		jdsw.PlayerName = playername
		jdsw.Type = ttone.Type
		jdsw.Time = ttone.Time
		jdsw.Count = ttone.Num
		jdsw.Src = "抽奖"
		a.SendMsgToAllClient("SC_ScrollWords", jdsw)
	}

}

//获取商店信息
func (turntab *TurnTableLogic) DoOneTurnTable(uid int, a *HallAgent) *datamsg.SC_DoTurnTable {

	log.Info("DoOneTurnTable")

	info := turntab.GetTurnTableInfo(uid)

	turntab.Lock.Lock()
	defer turntab.Lock.Unlock()
	if info.FreeTime > 0 {
		//检查砖石是否足够
		gold := 0
		db.DbOne.GetPlayerOneInfo(uid, "userbaseinfo", "gold", &gold)
		if gold < info.OnePrice {
			return nil
		}
		db.DbOne.SetPlayerOneInfo(uid, "userbaseinfo", "gold", gold-info.OnePrice)

	} else {
		day := time.Now().Format("2006-01-02 15:04:05")
		//设置免费获取时间
		db.DbOne.SetPlayerOneInfo(uid, "userbaseinfo", "lastturntabletime", day)

		info.FreeTime = 60 * 60 * 24
	}
	//随机道具
	randTurnTable := turntab.getRandTurnTable(info.TurnTables)
	//更改概率
	turntab.changeValue(info.TurnTables, randTurnTable)
	for k, v := range info.TurnTables {
		info.TurnTables[k].Value = turntab.IdValue.Get(v.Id).(int)
	}

	//玩家获得道具
	if randTurnTable.Type < 1000 {
		GetItemManager().AddItemsTime(uid, randTurnTable.Type, randTurnTable.Num)
	} else {

		GetItemManager().AddItemsTime(uid, randTurnTable.Type, randTurnTable.Time)
	}
	turntab.NoticeAll(uid, randTurnTable, a)

	//金币
	gold1 := 0
	db.DbOne.GetPlayerOneInfo(uid, "userbaseinfo", "gold", &gold1)
	//
	ids := make([]int, 1)
	ids[0] = randTurnTable.Id
	jd := &datamsg.SC_DoTurnTable{*info, ids, gold1}

	return jd

}

//获取商店信息
func (turntab *TurnTableLogic) DoLookViewTurnTable(uid int, a *HallAgent) *datamsg.SC_DoTurnTable {

	log.Info("DoLookViewTurnTable")

	info := turntab.GetTurnTableInfo(uid)
	if info.LookViewRemainder <= 0 {
		return nil
	}

	turntab.Lock.Lock()
	defer turntab.Lock.Unlock()

	//增加观看视频抽奖次数
	db.DbOne.AddPlayerOneInfo(uid, "userbaseinfo", "lookviewcount", 1)
	info.LookViewRemainder--

	//随机道具
	randTurnTable := turntab.getRandTurnTable(info.TurnTables)
	//更改概率
	turntab.changeValue(info.TurnTables, randTurnTable)
	for k, v := range info.TurnTables {
		info.TurnTables[k].Value = turntab.IdValue.Get(v.Id).(int)
	}

	//玩家获得道具
	if randTurnTable.Type < 1000 {
		GetItemManager().AddItemsTime(uid, randTurnTable.Type, randTurnTable.Num)
	} else {

		GetItemManager().AddItemsTime(uid, randTurnTable.Type, randTurnTable.Time)
	}
	turntab.NoticeAll(uid, randTurnTable, a)

	//金币
	gold1 := 0
	db.DbOne.GetPlayerOneInfo(uid, "userbaseinfo", "gold", &gold1)
	//
	ids := make([]int, 1)
	ids[0] = randTurnTable.Id
	jd := &datamsg.SC_DoTurnTable{*info, ids, gold1}

	return jd

}

//获取商店信息
func (turntab *TurnTableLogic) DoTenTurnTable(uid int, a *HallAgent) *datamsg.SC_DoTurnTable {

	log.Info("DoTenTurnTable")

	info := turntab.GetTurnTableInfo(uid)

	turntab.Lock.Lock()
	defer turntab.Lock.Unlock()

	//检查砖石是否足够
	gold := 0
	db.DbOne.GetPlayerOneInfo(uid, "userbaseinfo", "gold", &gold)
	if gold < info.TenPrice {
		return nil
	}
	db.DbOne.SetPlayerOneInfo(uid, "userbaseinfo", "gold", gold-info.TenPrice)

	ids := make([]int, 10)
	for i := 0; i < 10; i++ {
		//随机道具
		randTurnTable := turntab.getRandTurnTable(info.TurnTables)
		//更改概率
		turntab.changeValue(info.TurnTables, randTurnTable)
		for k, v := range info.TurnTables {
			info.TurnTables[k].Value = turntab.IdValue.Get(v.Id).(int)
		}

		//玩家获得道具
		if randTurnTable.Type < 1000 {
			GetItemManager().AddItemsTime(uid, randTurnTable.Type, randTurnTable.Num)
		} else {
			GetItemManager().AddItemsTime(uid, randTurnTable.Type, randTurnTable.Time)
		}
		turntab.NoticeAll(uid, randTurnTable, a)

		ids[i] = randTurnTable.Id
	}

	//金币
	gold1 := 0
	db.DbOne.GetPlayerOneInfo(uid, "userbaseinfo", "gold", &gold1)
	//

	jd := &datamsg.SC_DoTurnTable{*info, ids, gold1}

	return jd

}

////抽奖
//type SC_DoTurnTable struct {
//	SC_TurnTableInfo
//	Ids []int //中奖的ID
//}

//type SC_TurnTableInfo struct {
//	conf.TurnTableConfig
//	FreeTime int //免费抽的倒计时 0表示可以免费抽了
//}
//type TurnTableConfig struct {
//	Version    string      //版本
//	TurnTables []TurnTable //商品
//	OnePrice   int         //单价
//	TenPrice   int         //10次的价格
//}
//type TurnTable struct {
//	Id            int //商品ID
//	Type          int //商品类型
//	Time          int ////持续时间 以天为单位
//	Num           int //数量
//	InitLuckValue int //初始幸运值
//	Add           int //增长幸运值
//	Value         int //当前值
//}
