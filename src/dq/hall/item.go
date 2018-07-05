package hall

import (
	"encoding/json"
	"time"
	//"dq/conf"
	"dq/datamsg"
	"dq/db"
	"dq/log"
	//"dq/timer"
	"dq/utils"
	"sync"
	//"time"
)

var (
	itemManager = &ItemManager{Players: utils.NewBeeMap(), newUserLock: new(sync.Mutex)}
	//锁
)

type ItemManager struct {
	Players     *utils.BeeMap
	newUserLock *sync.Mutex
}

func GetItemManager() *ItemManager {
	return itemManager
}

//type ItemInfo struct {
//	Type         int
//	ExpiryTime string //到期时间
//	IsExpiry   bool   //是否到期
//}
//type SC_ItemInfo struct {
//	Items []ItemInfo
//}

type PlayerItem struct {
	Uid   int
	Items []datamsg.ItemInfo

	lock *sync.RWMutex
}

func (player *PlayerItem) readDB() {
	player.lock.Lock()
	defer player.lock.Unlock()

	items := ""
	err := db.DbOne.GetPlayerOneInfo(player.Uid, "userbaseinfo", "items", &items)
	if err != nil {
		log.Info(err.Error())
		return
	}

	err = json.Unmarshal([]byte(items), &player.Items)
	if err != nil {
		log.Info(err.Error())
		return
	}

}
func (player *PlayerItem) writeDB() {
	player.lock.Lock()
	defer player.lock.Unlock()

	//player.Items = append(player.Items, datamsg.ItemInfo{Type: 122, ExpiryTime: "2006-01-02 15:11:11"})

	itemsStr, err := json.Marshal(player.Items)
	if err != nil {
		log.Info(err.Error())
		return
	}
	err = db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "items", itemsStr)
	if err != nil {
		log.Error(err.Error())
		return
	}
}

//Beiyongtime     int
//	Steptime        int
func (player *PlayerItem) getItemsInfo() *datamsg.SC_ItemInfo {
	player.lock.RLock()
	defer player.lock.RUnlock()

	jd := &datamsg.SC_ItemInfo{}
	db.DbOne.GetPlayerManyInfo(player.Uid, "userbaseinfo", "firstqizi,secondqizi,qizi_move,qizi_move_trail,qizi_floor,qizi_lastplay,beiyongtime,steptime",
		&jd.Firstqizi, &jd.Secondqizi, &jd.Qizi_move, &jd.Qizi_move_trail, &jd.Qizi_floor, &jd.Qizi_lastplay, &jd.Beiyongtime, &jd.Steptime)

	jd.Items = make([]datamsg.ItemInfo, len(player.Items))

	for k, v := range player.Items {

		jd.Items[k] = v
		//CheckTimeIsExpiry
		jd.Items[k].IsExpiry = utils.CheckTimeIsExpiry(jd.Items[k].ExpiryTime)

		log.Info("-getItemsInfo--type:%d----time:%s", v.Type, v.ExpiryTime)

	}
	return jd
}

//
func (player *PlayerItem) getItemIndex(itemType int) int {
	//player.lock.RLock()
	//defer player.lock.RUnlock()
	//log.Info("len:%d", len(player.Items))

	for k, v := range player.Items {
		//log.Info("type:%d---itemtype:%d---k:%d", v.Type, itemType, k)
		if v.Type == itemType {
			return k
		}
	}

	return -1
}

//增加道具时间(包括购买道具和赠送道具) 如果type为1 后面的day则为数量
func (player *PlayerItem) AddItemsTime(itemType int, days int) {
	player.lock.Lock()
	defer player.lock.Unlock()

	//log.Info("AddItemsTime")
	itemIndex := player.getItemIndex(itemType)
	if itemIndex >= 0 {
		item := &player.Items[itemIndex]
		//过期
		if utils.CheckTimeIsExpiry(item.ExpiryTime) {

			item.ExpiryTime = time.Now().AddDate(0, 0, days).Format("2006-01-02 15:04:05")
			log.Info("-----11-----fdsfa")
		} else {
			t1, _ := time.Parse("2006-01-02 15:04:05", item.ExpiryTime)
			item.ExpiryTime = t1.AddDate(0, 0, days).Format("2006-01-02 15:04:05")
			//log.Info("-----22-----fdsfa")
		}
	} else {
		item := datamsg.ItemInfo{Type: itemType}
		item.ExpiryTime = time.Now().AddDate(0, 0, days).Format("2006-01-02 15:04:05")
		player.Items = append(player.Items, item)
		//log.Info("-----33-----fdsfa")

	}

	player.SetUseItemType(itemType)
	//player.writeDB()

}

//检查单个道具是否过期 true 表示过期
func (player *PlayerItem) CheckOneItemTime(itemType int) bool {
	itemIndex := player.getItemIndex(itemType)
	if itemIndex >= 0 {
		item := &player.Items[itemIndex]
		//过期
		if utils.CheckTimeIsExpiry(item.ExpiryTime) {

			return true
		} else {
			return false

		}
	} else {
		return true

	}
}

//设置当前使用特效类型
func (player *PlayerItem) SetUseItemType(itemType int) bool {
	if itemType < 1000 {
		return false
	}
	//log.Info("-----44-----fdsfa")
	//检查是否过期
	if player.CheckOneItemTime(itemType) {
		return false
	}
	//log.Info("-----55-----fdsfa")

	if itemType < 1100 {
		firstqizi := 1001
		db.DbOne.GetPlayerOneInfo(player.Uid, "userbaseinfo", "firstqizi", &firstqizi)
		if firstqizi != itemType {
			db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "firstqizi", itemType)
			db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "secondqizi", firstqizi)
		}

	} else if itemType < 1200 {
		db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "qizi_move", itemType)
	} else if itemType < 1300 {
		db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "qizi_move_trail", itemType)
	} else if itemType < 1400 {
		db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "qizi_floor", itemType)
	} else if itemType < 1500 {
		db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "qizi_lastplay", itemType)
	} else if itemType < 1600 {
		db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "beiyongtime", itemType)
	} else if itemType < 1700 {
		db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "steptime", itemType)
	}
	return true
}

//检查道具时间
func (player *PlayerItem) CheckItemsTime() {
	player.lock.Lock()
	defer player.lock.Unlock()

	firstqizi := 0
	secondqizi := 0
	qizi_move := 0
	qizi_move_trail := 0
	qizi_floor := 0
	qizi_lastplay := 0
	beiyongtime := 0
	steptime := 0

	db.DbOne.GetPlayerManyInfo(player.Uid, "userbaseinfo", "firstqizi,secondqizi,qizi_move,qizi_move_trail,qizi_floor,qizi_lastplay,beiyongtime,steptime",
		&firstqizi, &secondqizi, &qizi_move, &qizi_move_trail, &qizi_floor, &qizi_lastplay, &beiyongtime, &steptime)

	//log.Info("--%d-%d-%d-%d-%d-%d-", firstqizi, secondqizi, qizi_move, qizi_move_trail, qizi_floor, qizi_lastplay)
	//itemIndex := player.getItemIndex(firstqizi)
	if player.CheckOneItemTime(firstqizi) && firstqizi != 1001 {
		db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "firstqizi", 1001)
		//log.Info("fdsafdsa")

	}
	if player.CheckOneItemTime(secondqizi) && secondqizi != 1002 {
		db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "secondqizi", 1002)

	}
	if player.CheckOneItemTime(qizi_move) && qizi_move != 0 {
		db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "qizi_move", 0)
	}
	if player.CheckOneItemTime(qizi_move_trail) && qizi_move_trail != 0 {
		db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "qizi_move_trail", 0)
	}
	if player.CheckOneItemTime(qizi_floor) && qizi_floor != 0 {
		db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "qizi_floor", 0)
	}
	if player.CheckOneItemTime(qizi_lastplay) && qizi_lastplay != 0 {
		db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "qizi_lastplay", 0)
	}
	if player.CheckOneItemTime(beiyongtime) && beiyongtime != 0 {
		db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "beiyongtime", 0)
	}
	if player.CheckOneItemTime(steptime) && steptime != 0 {
		db.DbOne.SetPlayerOneInfo(player.Uid, "userbaseinfo", "steptime", 0)
	}

}

//获取玩家
func (itemManager *ItemManager) GetPlayer(uid int) *PlayerItem {
	if itemManager.Players.Check(uid) == false {
		itemManager.newUserLock.Lock()
		defer itemManager.newUserLock.Unlock()

		if itemManager.Players.Check(uid) == false {
			player := &PlayerItem{Uid: uid, lock: new(sync.RWMutex)}
			player.readDB()
			itemManager.Players.Set(uid, player)
			player.CheckItemsTime()
			return itemManager.Players.Get(uid).(*PlayerItem)
		} else {
			//itemManager.Players.Get(uid).(*PlayerItem).CheckItemsTime()
			return itemManager.Players.Get(uid).(*PlayerItem)
		}
	}
	//itemManager.Players.Get(uid).(*PlayerItem).CheckItemsTime()
	return itemManager.Players.Get(uid).(*PlayerItem)
}

//删除玩家
func (itemManager *ItemManager) DeletePlayer(uid int) {
	if itemManager.Players.Check(uid) == true {
		itemManager.newUserLock.Lock()
		defer itemManager.newUserLock.Unlock()

		if itemManager.Players.Check(uid) == true {
			player := itemManager.Players.Get(uid).(*PlayerItem)
			if player != nil {
				player.writeDB()
				itemManager.Players.Delete(uid)
			}
		}
	}
}

//deleteAll
func (itemManager *ItemManager) DeleteAll() {

	itemManager.newUserLock.Lock()
	defer itemManager.newUserLock.Unlock()
	items := itemManager.Players.Items()
	for _, v := range items {
		player := v.(*PlayerItem)
		if player != nil {
			player.writeDB()
			//itemManager.Players.Delete(uid)
		}
	}
	itemManager.Players.DeleteAll()

}

//获取玩家道具信息
func (itemManager *ItemManager) GetItemsInfo(uid int) *datamsg.SC_ItemInfo {

	player := itemManager.GetPlayer(uid)
	if player != nil {
		//player.AddItemsTime(1508, 7)
		return player.getItemsInfo()
	}
	return nil
}

//获取玩家道具信息 如果type为1 后面的day则为数量
func (itemManager *ItemManager) AddItemsTime(uid int, itemtype int, day int) {

	//砖石
	if itemtype == 1 {
		db.DbOne.AddPlayerOneInfo(uid, "userbaseinfo", "gold", day)
		return
	}

	player := itemManager.GetPlayer(uid)
	if player != nil {
		player.AddItemsTime(itemtype, day)
		player.writeDB()
	}
}

//设置当前使用特效类型
func (itemManager *ItemManager) SetUseItemType(uid int, itemType int) bool {

	//type SC_UpdateUsedItem struct {
	//	Firstqizi       int
	//	Secondqizi      int
	//	Qizi_move       int
	//	Qizi_move_trail int
	//	Qizi_floor      int
	//	Qizi_lastplay   int
	//}
	player := itemManager.GetPlayer(uid)
	if player != nil {
		//player.AddItemsTime(1508, 7)
		return player.SetUseItemType(itemType)
	}
	return false
}
