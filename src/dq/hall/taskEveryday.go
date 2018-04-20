package hall

import (
	"dq/conf"
	"dq/db"
	"dq/log"
	"dq/utils"
	"sync"
)

var (
	taskEveryday = &TaskEveryday{PlayerTskEd: utils.NewBeeMap()}
	//锁
	lock = new(sync.RWMutex)
)

type TaskEveryday struct {
	PlayerTskEd *utils.BeeMap
}

func GetTaskEveryday() *TaskEveryday {
	return taskEveryday
}

type UserTaskEveryday struct {
	Uid     int
	Date    string
	DBValue *utils.BeeMap //数据库中的字段和值
}

func (user *UserTaskEveryday) readValueFromDB() {
	//从数据库获取值
	if err := db.DbOne.GetPlayerTaskEd(user.Uid, &user.Date, user.DBValue); err != nil {
		log.Info(err.Error())
	}
	items := user.DBValue.Items()
	for k, v := range items {
		log.Info("---k:%s---value:%v", k, v)
	}
	log.Info("---day:" + user.Date)

}

func (taskE *TaskEveryday) getPlayer(uid int) *UserTaskEveryday {
	if taskE.PlayerTskEd.Check(uid) == false {
		player := &UserTaskEveryday{}
		//taskE.PlayerTskEd.Set(uid, player)
		player.DBValue = utils.NewBeeMap()
		player.Uid = uid

		//获取值
		cfg := conf.GetTaskEveryDayCfg()
		for _, v := range cfg.Task {
			player.DBValue.Set(v.GetTagDBFieldName, 0)
			player.DBValue.Set(v.ProgressDBFieldName, v.InitialValue)
		}
		player.readValueFromDB()

		taskE.PlayerTskEd.Set(uid, player)

	}

	return (taskE.PlayerTskEd.Get(uid)).(*UserTaskEveryday)
}

//获取用户每日任务 完成的切没有领取奖励的任务个数
func (taskE *TaskEveryday) getCompleteNumOfTskEd(uid int) {
	log.Info("----getCompleteNumOfTskEd----")
	//t1 := time.Now().Format("2006-01-02")
	//	type TaskEveryDayConfig struct {
	//	Version int //版本
	//	Task    []TaskConfig
	//}
	taskE.getPlayer(uid)
	//cfg := conf.GetTaskEveryDayCfg()
	//	for v := range cfg.Task {

	//	}

	if false {

	}
}
