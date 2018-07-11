package hall

import (
	"dq/datamsg"
	"dq/db"
	"dq/log"
	"dq/utils"
	"sync"
)

var (
	friends          = &Friends{Lock: new(sync.Mutex)}
	friendsMaxCount  = 200
	friendsShowCount = 15
)

type Friends struct {
	Lock *sync.Mutex

	//用户锁
	UserLock [1000]*sync.Mutex
}

func GetFriends() *Friends {
	return friends
}

//更新相互战绩
func (friends *Friends) UpdateRecord(winuid int, loseuid int) {
	if winuid == loseuid {
		return
	}

	db.DbOne.UpdateRecord(winuid, loseuid)
}

//互加为好友
func (friends *Friends) AddFriends(uid1 int, uid2 int) {
	if uid1 == uid2 {
		return
	}

	user1friends := ""
	err := db.DbOne.GetPlayerOneInfo(uid1, "userbaseinfo", "friends_id", &user1friends)
	if err != nil {
		//log.Error("AddFriends-")
		return
	}
	user2friends := ""
	err = db.DbOne.GetPlayerOneInfo(uid2, "userbaseinfo", "friends_id", &user2friends)
	if err != nil {
		log.Info("user2friends--%s", user2friends)
		return
	}

	ids1 := utils.SplitStringToIntMap(user1friends)
	ids2 := utils.SplitStringToIntMap(user2friends)

	//log.Info("AddFriends--id1len:%d---id2:len:%d", len(ids1), len(ids2))
	//有一方好友达到上限就不能互加为好友了
	if len(ids1) > friendsMaxCount || len(ids2) > friendsMaxCount {
		return
	}
	//已经是好友了
	if ids1[uid2] == true || ids2[uid1] == true {
		return
	}

	//
	db.DbOne.AddToFriends(uid1, uid2)

}

//获取好友信息
func (friends *Friends) getFriendsInfo(uid int, count int, a *HallAgent) *datamsg.SC_FriendsInfo {

	jd := &datamsg.SC_FriendsInfo{}
	if count <= 0 {
		return jd
	}

	//type FriendInfo struct {
	//	Uid       int
	//	Name      string
	//	Avatar    string
	//	Score     int
	//	State     int //0离线 1大厅 2比赛中 3观战中
	//	FriendWin int //朋友赢得次数
	//	MyWin     int //我赢得次数
	//	GiveState int //打赏状态 0表示还没打赏 1表示已经打赏

	//}

	jd.Friends = make([]datamsg.FriendInfo, 0)
	userfriends := ""
	err := db.DbOne.GetPlayerOneInfo(uid, "userbaseinfo", "friends_id", &userfriends)
	if err != nil || userfriends == "" {
		return jd
	}
	ids := utils.SplitStringToIntMap(userfriends)

	showids := make([]int, 0)
	if len(ids) > 0 {

		//优先查找在线玩家
		for k, _ := range ids {
			onlinestate := a.PlayerGameState.Get(k)
			if onlinestate != nil {
				showids = append(showids, k)
				delete(ids, k)
			}
		}
		if len(showids) < friendsShowCount {
			for k, _ := range ids {
				showids = append(showids, k)
				if len(showids) >= friendsShowCount {
					break
				}
			}
		}
		//		for _, v := range showids {
		//			log.Info("----showids uid:%d--", v)
		//		}

		//获取用户基本信息
		db.DbOne.GetFriendsBaseInfo(showids, &jd.Friends)

		//获取对战信息
		db.DbOne.GetFriendsBattleInfo(uid, &jd.Friends)

		for k, v := range jd.Friends {
			//用户在线状态信息
			onlinestate := a.PlayerGameState.Get(v.Uid)
			//log.Info("----friend uid:%d--", v.Uid)
			if onlinestate == nil {
				jd.Friends[k].State = 0
			} else {
				jd.Friends[k].State = onlinestate.(int)
			}

			//			log.Info("----friend uid:%d-- name:%s--Seasonscore:%d--mywin:%d---friendwin:%d",
			//				v.Uid, v.Name, v.Seasonscore, v.MyWin, v.FriendWin)
		}
	}

	return jd

}

func (friends *Friends) Init() {
	friends.Lock.Lock()
	defer friends.Lock.Unlock()

	//	for i := 0; i < 1000; i++ {
	//		friends.UserLock[i] = new(sync.Mutex)
	//	}

}
