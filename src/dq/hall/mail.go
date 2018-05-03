package hall

import (
	"dq/conf"
	"dq/datamsg"
	"dq/db"
	"dq/log"
	"dq/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	mail         = &Mail{Lock: new(sync.Mutex)}
	mailMaxCount = 10
)

type Mail struct {
	PublicMail *utils.BeeMap
	MaxIndex   int
	Lock       *sync.Mutex

	//用户锁
	UserLock [1000]*sync.Mutex
}

func GetMail() *Mail {
	return mail
}

//type RewardsConfig struct {
//	Type  int
//	Count int
//	Time  int
//}

//type MailInfo struct {
//	Id        int
//	SendName  string
//	Title     string
//	Content   string
//	RecUid    int
//	Date      string
//	Reward    []conf.RewardsConfig
//	ReadState int
//	GetState  int
//}
func (mail *Mail) DoPresenterMail(PresenterUid int, content string, goldCount int) {

	mailp := &datamsg.MailInfo{}
	mailp.SendName = "宝石五指棋"
	mailp.Title = "推荐有奖"
	mailp.Content = content
	mailp.RecUid = PresenterUid
	mailp.Date = time.Now().Format("2006-01-02")
	tt := conf.RewardsConfig{Type: 1, Count: goldCount, Time: -1}
	mailp.Reward = append(mailp.Reward, tt)
	mailp.ReadState = 0
	mailp.GetState = 0

	db.DbOne.WritePrivateMailInfo(PresenterUid, mailp)
}

func (mail *Mail) CheckUserPublicMail(uid int) {
	//mail.Lock.Lock()
	//defer mail.Lock.Unlock()

	mails := ""
	publicIndex := 0
	db.DbOne.GetPlayerOtherInfo(uid, &mails, &publicIndex)

	if mail.MaxIndex > publicIndex {
		//items := mail.PublicMail.Items()
		for i := publicIndex + 1; i <= mail.MaxIndex; i++ {
			item := mail.PublicMail.Get(i)
			if item != nil {
				log.Info("---------insert:%d-----%v", i, item.(*datamsg.MailInfo))

				db.DbOne.WritePrivateMailInfoFromPublic(uid, item.(*datamsg.MailInfo), i)
			}

		}
	}
}

//邮件超过50封 就自动获取奖励并且删除邮件
func (mail *Mail) CheckUserMail(uid int) {
	log.Info("------CheckUserMail----")
	usermails := ""
	err := db.DbOne.GetPlayerOneInfo(uid, "userbaseinfo", "mails_id", &usermails)
	if err != nil {
		log.Info(err.Error())
		return
	}
	if usermails == "" {
		log.Info("usermails == null")
		return
	}

	log.Info("------usermails---:" + usermails)
	allmails := strings.Split(usermails, ",")
	log.Info("------usermails--count-:%d", len(allmails))
	if len(allmails) > mailMaxCount {
		for i := 0; i < len(allmails)-mailMaxCount; i++ {
			mailid, err := strconv.Atoi(allmails[i])
			if err != nil {
				log.Info(err.Error())
				continue
			}
			log.Info("----id:%d", mailid)
			if mailid <= 0 {

				continue
			}

			//db.DbOne.GetMailRewards(user.Uid, v.Rewards, mailid)
			mail.getMailRewards(uid, mailid)

		}

		newmailstr := ""
		for i := len(allmails) - mailMaxCount; i < len(allmails); i++ {

			newmailstr = newmailstr + allmails[i]

		}
		db.DbOne.SetPlayerOneInfo(uid, "userbaseinfo", "mails_id", newmailstr)

	}

}

//获取未读或未领取邮件数量(新邮件)
func (mail *Mail) getNewMailNum(uid int) int {
	mailnum := 0

	usermails := ""
	err := db.DbOne.GetPlayerOneInfo(uid, "userbaseinfo", "mails_id", usermails)
	if err != nil || usermails == "" {
		return mailnum
	}
	allmails := strings.Split(usermails, ",")
	if len(allmails) > 0 {
		for i := 0; i < len(allmails)-mailMaxCount; i++ {
			mailid, err := strconv.Atoi(allmails[i])
			if err != nil {
				continue
			}
			if mailid <= 0 {
				continue
			}
			isget := -1
			err = db.DbOne.GetPlayerOneInfo(mailid, "mail", "getstate", isget)
			if isget == 0 {
				mailnum++
				return mailnum
			}

			//mail.getMailRewards(uid, mailid)

		}

	}
	//db.DbOne.GetMailRewards(uid, mailid)

	return mailnum

}

//获取任务奖励
func (mail *Mail) getMailRewards(uid int, mailid int) {
	//user.lock.RLock()
	//defer user.lock.RUnlock()
	lockid := uid % 1000
	mail.UserLock[lockid].Lock()
	defer mail.UserLock[lockid].Unlock()

	db.DbOne.GetMailRewards(uid, mailid)

}

//获取邮件信息
func (mail *Mail) getMailInfo(uid int, count int) *datamsg.SC_MailInfo {
	//	type MailInfo struct {
	//	Id        int
	//	SendName  string
	//	Title     string
	//	Content   string
	//	RecUid    int
	//	Date      string
	//	Reward    []conf.RewardsConfig
	//	ReadState int
	//	GetState  int
	//}

	jd := &datamsg.SC_MailInfo{}
	if count <= 0 {
		return jd
	}

	jd.Mails = make([]datamsg.MailInfo, count)
	usermails := ""
	err := db.DbOne.GetPlayerOneInfo(uid, "userbaseinfo", "mails_id", usermails)
	if err != nil || usermails == "" {
		return jd
	}
	allmails := strings.Split(usermails, ",")
	if len(allmails) > 0 {
		startindex := 0
		if len(allmails) > count {
			startindex = len(allmails) - count
		}
		index := 0
		for i := startindex; i < len(allmails); i++ {
			mailid, err := strconv.Atoi(allmails[i])
			if err != nil {
				index++
				continue
			}
			if mailid <= 0 {
				index++
				continue
			}
			jd.Mails[index] = datamsg.MailInfo{}
			index++

		}

	}

	return jd

	//	for k, v := range cfg.Task {

	//		jd.Task[k] = datamsg.MsgTaskInfo{}
	//		jd.Task[k].Id = v.Id
	//		jd.Task[k].Type = v.Type
	//		jd.Task[k].DestValue = v.DestValue
	//		jd.Task[k].ProgressValue = 0
	//		jd.Task[k].State = 0
	//		jd.Task[k].Rewards = v.Rewards
	//		value := user.DBValue.Get(v.ProgressDBFieldName)
	//		get := user.DBValue.Get(v.GetTagDBFieldName)
	//		if value != nil && get != nil {
	//			jd.Task[k].ProgressValue = value.(int)

	//			if value.(int) >= v.DestValue && get.(int) == 1 {
	//				jd.Task[k].State = 2
	//				jd.Task[k].ProgressValue = v.DestValue
	//			} else if value.(int) >= v.DestValue && get.(int) != 1 {
	//				jd.Task[k].State = 1
	//				jd.Task[k].ProgressValue = v.DestValue
	//			} else if value.(int) < v.DestValue {
	//				jd.Task[k].State = 0
	//			}
	//		}

	//	}
	//	return jd

}

func (mail *Mail) Init() {
	mail.Lock.Lock()
	defer mail.Lock.Unlock()

	mail.PublicMail = utils.NewBeeMap()
	mail.MaxIndex = 0
	for i := 0; i < 1000; i++ {
		mail.UserLock[i] = new(sync.Mutex)
	}

	db.DbOne.GetPublicMailInfo(mail.PublicMail, &mail.MaxIndex)
}
