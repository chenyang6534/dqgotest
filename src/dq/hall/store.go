package hall

import (
	"dq/conf"
	"dq/datamsg"
	"dq/db"
	"dq/log"
	//"dq/utils"
	//"strconv"
	//"strings"
	"sync"
	"time"
)

var (
	store = &Store{Lock: new(sync.Mutex)}
)

type Store struct {
	Lock *sync.Mutex

	//用户锁
	UserLock [1000]*sync.Mutex
}

func GetStore() *Store {
	return store
}

//type StoreInfo struct {
//	conf.Commodity
//	IsStartSale bool
//}

////商店信息
//type SC_StoreInfo struct {
//	Commoditys []StoreInfo
//}

func (store *Store) Buy(uid int, commodityid int, index int) bool {

	log.Info("buy1")
	storecfg := conf.GetStoreConfig()
	if index < 0 || index >= len(storecfg.Commoditys) {
		return false
	}
	log.Info("buy2")
	if storecfg.Commoditys[index].Id != commodityid {
		return false
	}
	log.Info("buy3")
	price := storecfg.Commoditys[index].SalePrice * (storecfg.Commoditys[index].SaleDiscount / 10)
	//itemType := storecfg.Commoditys[index].Type
	log.Info("price:%d", price)
	gold := 0
	db.DbOne.GetPlayerOneInfo(uid, "userbaseinfo", "gold", &gold)
	log.Info("gold:%d", gold)
	if gold >= price {
		//购买成功
		db.DbOne.SetPlayerOneInfo(uid, "userbaseinfo", "gold", gold-price)
		GetItemManager().AddItemsTime(uid, storecfg.Commoditys[index].Type, storecfg.Commoditys[index].Time)
		return true
	}

	return false
}

//获取商店信息
func (store *Store) getStoreInfo() *datamsg.SC_StoreInfo {

	jd := &datamsg.SC_StoreInfo{}
	storecfg := conf.GetStoreConfig()
	if len(storecfg.Commoditys) <= 0 {
		return jd
	}

	jd.Commoditys = make([]datamsg.StoreInfo, len(storecfg.Commoditys))
	nowtime, _ := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"))
	for k, v := range storecfg.Commoditys {
		jd.Commoditys[k] = datamsg.StoreInfo{v, false}

		start, _ := time.Parse("2006-01-02 15:04:05", v.StartTime)
		end, _ := time.Parse("2006-01-02 15:04:05", v.EndTime)

		if start.Before(nowtime) && end.After(nowtime) {
			jd.Commoditys[k].IsStartSale = true
		} else {
			jd.Commoditys[k].IsStartSale = false
		}
		log.Info("---getStoreInfo------id:%d----isstartsale:%t-", jd.Commoditys[k].Id, jd.Commoditys[k].IsStartSale)
	}

	return jd

}
