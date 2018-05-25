package hall

import (
	"dq/conf"
	"dq/datamsg"
	"dq/db"
	"dq/log"
	//"dq/timer"
	"dq/utils"
	"sync"
	"time"
)

var (
	taskEveryday = &TaskEveryday{PlayerTskEd: utils.NewBeeMap(), newUserLock: new(sync.Mutex)}
	//锁
	lock = new(sync.RWMutex)
)

type TaskEveryday struct {
	PlayerTskEd *utils.BeeMap
	newUserLock *sync.Mutex
}

func GetTaskEveryday() *TaskEveryday {
	return taskEveryday
}

type UserTaskEveryday struct {
	Uid     int
	Version string
	Date    string
	DBValue *utils.BeeMap //数据库中的字段和值
	lock    *sync.RWMutex
}

func (user *UserTaskEveryday) readValueFromDB() {
	//从数据库获取值
	if err := db.DbOne.GetPlayerTaskEd(user.Uid, &user.Date, user.DBValue); err != nil {
		log.Info(err.Error())
	}
	//	items := user.DBValue.Items()
	//	for k, v := range items {
	//		log.Info("---k:%s---value:%v", k, v)
	//	}
	//log.Info("---day:" + user.Date)

}
func (user *UserTaskEveryday) doShare() {
	user.lock.RLock()
	defer user.lock.RUnlock()

	//以后优化
	user.DBValue.AddInt("share_count", 1)
	user.writeToDB()
}

//获取任务奖励
func (user *UserTaskEveryday) getTskEdRewards(taskid int) bool {
	user.lock.RLock()
	defer user.lock.RUnlock()

	//以后优化
	cfg := conf.GetTaskEveryDayCfg()
	for _, v := range cfg.Task {
		if v.Id == taskid {
			value := user.DBValue.Get(v.ProgressDBFieldName)
			get := user.DBValue.Get(v.GetTagDBFieldName)

			log.Info("---------value:%d----get:%d", value.(int), get.(int))
			if value != nil && get != nil {

				if value.(int) >= v.DestValue && get.(int) != 1 {
					user.DBValue.Set(v.GetTagDBFieldName, 1)
					//奖励
					db.DbOne.GetTaskRewards(user.Uid, v.Rewards, v.GetTagDBFieldName)

					return true
				}

			}
			return false
		}
	}

	return false
}

//获取任务信息
func (user *UserTaskEveryday) getTskEdInfo() *datamsg.SC_TskEdInfo {
	user.lock.RLock()
	defer user.lock.RUnlock()

	jd := &datamsg.SC_TskEdInfo{}

	cfg := conf.GetTaskEveryDayCfg()
	jd.Task = make([]datamsg.MsgTaskInfo, len(cfg.Task))
	for k, v := range cfg.Task {

		jd.Task[k] = datamsg.MsgTaskInfo{}
		jd.Task[k].Id = v.Id
		jd.Task[k].Type = v.Type
		jd.Task[k].DestValue = v.DestValue
		jd.Task[k].ProgressValue = 0
		jd.Task[k].State = 0
		jd.Task[k].Rewards = v.Rewards
		value := user.DBValue.Get(v.ProgressDBFieldName)
		get := user.DBValue.Get(v.GetTagDBFieldName)
		if value != nil && get != nil {
			jd.Task[k].ProgressValue = value.(int)

			if value.(int) >= v.DestValue && get.(int) == 1 {
				jd.Task[k].State = 2
				jd.Task[k].ProgressValue = v.DestValue
			} else if value.(int) >= v.DestValue && get.(int) != 1 {
				jd.Task[k].State = 1
				jd.Task[k].ProgressValue = v.DestValue
			} else if value.(int) < v.DestValue {
				jd.Task[k].State = 0
			}
		}

	}
	return jd

	//	type MsgTaskInfo struct {
	//	Id        int
	//	Type      int
	//	DestValue int //目标
	//	ProgressValue int //进度
	//	State		int //状态 0表示未完成 1表示已完成没领取 2表示已领取

	//}
	////每日任务信息
	//type SC_TskEdInfo() struct {
	//	Task	[]MsgTaskInfo
	//}

}

//获取完成任务且没领取奖励的任务个数
func (user *UserTaskEveryday) getCompleteNumOfTskEd() int {
	user.lock.RLock()
	defer user.lock.RUnlock()
	//获取值
	count := 0
	cfg := conf.GetTaskEveryDayCfg()
	for _, v := range cfg.Task {

		value := user.DBValue.Get(v.ProgressDBFieldName)
		get := user.DBValue.Get(v.GetTagDBFieldName)
		if value != nil && get != nil {
			if value.(int) >= v.DestValue && get.(int) != 1 {
				count++
			}
		}

	}
	return count

}

//写到数据库
func (user *UserTaskEveryday) writeToDB() {

	//	items := user.DBValue.Items()
	//	for k, _ := range items {
	//		user.DBValue.Set(k, 100)
	//	}

	db.DbOne.SetPlayerTaskEd(user.Uid, user.Date, user.DBValue)
}

func (user *UserTaskEveryday) doCheck() bool {
	user.lock.Lock()
	defer user.lock.Unlock()

	today := time.Now().Format("2006-01-02")
	if today != user.Date {

		items := user.DBValue.Items()
		for k, _ := range items {
			user.DBValue.Set(k, 0)
		}
		user.Date = today
		user.writeToDB()
		//
		return false
	}
	cfg := conf.GetTaskEveryDayCfg()
	if cfg.Version != user.Version {

		t1 := utils.NewBeeMap()

		//获取值
		cfg := conf.GetTaskEveryDayCfg()
		for _, v := range cfg.Task {
			t1.Set(v.GetTagDBFieldName, 0)
			t1.Set(v.ProgressDBFieldName, v.InitialValue)
		}
		user.Version = cfg.Version

		items := t1.Items()
		for k, _ := range items {
			if user.DBValue.Check(k) == true {
				t1.Set(k, user.DBValue.Get(k))
			}

		}
		user.DBValue = t1
		user.writeToDB()

		return false
	}

	return true
}

func (taskE *TaskEveryday) DeleteUserTaskEveryday(uid1 interface{}) {

	taskE.newUserLock.Lock()
	defer taskE.newUserLock.Unlock()

	uid := uid1.(int)

	if taskE.PlayerTskEd.Check(uid) == true {
		player := taskE.PlayerTskEd.Get(uid)

		player.(*UserTaskEveryday).writeToDB()
		taskE.PlayerTskEd.Delete(uid)
	}

}
func (taskE *TaskEveryday) DeleteAll() {

	taskE.newUserLock.Lock()
	defer taskE.newUserLock.Unlock()
	items := taskE.PlayerTskEd.Items()
	for _, v := range items {
		player := v.(*UserTaskEveryday)
		if player != nil {
			player.writeToDB()
			//itemManager.Players.Delete(uid)
		}
	}
	taskE.PlayerTskEd.DeleteAll()

	//	uid := uid1.(int)

	//	if taskE.PlayerTskEd.Check(uid) == true {
	//		player := taskE.PlayerTskEd.Get(uid)

	//		player.(*UserTaskEveryday).writeToDB()
	//		taskE.PlayerTskEd.Delete(uid)
	//	}

}

func (taskE *TaskEveryday) newUserTaskEveryday(uid int) {

	taskE.newUserLock.Lock()
	defer taskE.newUserLock.Unlock()

	if taskE.PlayerTskEd.Check(uid) == false {
		player := &UserTaskEveryday{}
		//taskE.PlayerTskEd.Set(uid, player)
		player.DBValue = utils.NewBeeMap()
		player.Uid = uid
		player.lock = new(sync.RWMutex)

		//获取值
		cfg := conf.GetTaskEveryDayCfg()
		for _, v := range cfg.Task {
			player.DBValue.Set(v.GetTagDBFieldName, 0)
			player.DBValue.Set(v.ProgressDBFieldName, v.InitialValue)
		}
		player.Version = cfg.Version

		//再次检查

		player.readValueFromDB()
		taskE.PlayerTskEd.Set(uid, player)
		//return player

		//timer.AddCallback(time.Second*60*30, taskE.DeleteUserTaskEveryday, uid)
	}

	//return player
}

func (taskE *TaskEveryday) getPlayer(uid int) *UserTaskEveryday {
	if taskE.PlayerTskEd.Check(uid) == false {

		taskE.newUserTaskEveryday(uid)

	}
	player := (taskE.PlayerTskEd.Get(uid)).(*UserTaskEveryday)

	//检查日期是否过期 版本是否过期
	player.doCheck()

	return player
}

//获取用户每日任务 完成的切没有领取奖励的任务个数
func (taskE *TaskEveryday) getCompleteNumOfTskEd(uid int) int {
	log.Info("----getCompleteNumOfTskEd----")
	//t1 := time.Now().Format("2006-01-02")
	//	type TaskEveryDayConfig struct {
	//	Version int //版本
	//	Task    []TaskConfig
	//}
	player := taskE.getPlayer(uid)
	if player != nil {
		return player.getCompleteNumOfTskEd()
	}
	return 0
}

func (taskE *TaskEveryday) getTskEdInfo(uid int) *datamsg.SC_TskEdInfo {
	log.Info("----getTskEdInfo----")
	//t1 := time.Now().Format("2006-01-02")
	//	type TaskEveryDayConfig struct {
	//	Version int //版本
	//	Task    []TaskConfig
	//}
	player := taskE.getPlayer(uid)
	if player != nil {
		return player.getTskEdInfo()
	}
	return nil
}

func (taskE *TaskEveryday) getTskEdRewards(uid int, taskid int) bool {
	log.Info("----getTskEdRewards----")
	//t1 := time.Now().Format("2006-01-02")
	//	type TaskEveryDayConfig struct {
	//	Version int //版本
	//	Task    []TaskConfig
	//}
	player := taskE.getPlayer(uid)
	if player != nil {
		return player.getTskEdRewards(taskid)
	}
	return false
}

func (taskE *TaskEveryday) doShare(uid int) {

	player := taskE.getPlayer(uid)
	if player != nil {
		player.doShare()
	}
}

//赛季比赛赢
func (taskE *TaskEveryday) SeasonMatchWin(uid int) {
	player := taskE.getPlayer(uid)
	if player != nil {
		player.DBValue.AddInt("sm_win_count", 1)
	}
}

//好友间比赛赢
func (taskE *TaskEveryday) FriendMatchWin(uid int) {
	player := taskE.getPlayer(uid)
	if player != nil {
		player.DBValue.AddInt("fm_win_count", 1)
	}
}

//赢
func (taskE *TaskEveryday) Win(uid int) {
	player := taskE.getPlayer(uid)
	if player != nil {
		player.DBValue.AddInt("win_count", 1)
	}
}

//赛季比赛完成一局
func (taskE *TaskEveryday) SeasonMatchPlay(uid int) {
	player := taskE.getPlayer(uid)
	if player != nil {
		player.DBValue.AddInt("sm_play_count", 1)
	}
}

//好友间比赛完成一局
func (taskE *TaskEveryday) FriendMatchPlay(uid int) {
	player := taskE.getPlayer(uid)
	if player != nil {
		player.DBValue.AddInt("fm_play_count", 1)
	}
}

//完成一局
func (taskE *TaskEveryday) Play(uid int) {
	player := taskE.getPlayer(uid)
	if player != nil {
		player.DBValue.AddInt("play_count", 1)
	}
}
