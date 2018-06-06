package hall

import (
	"dq/conf"
	"dq/datamsg"
	"dq/db"
	"dq/log"
	"dq/utils"
	"encoding/json"
	"strconv"
	//"strings"
	"sync"
	"time"
)

var (
	mail         = &Mail{Lock: new(sync.Mutex)}
	mailMaxCount = 3
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

	allmails := utils.SplitStringToIntArray(usermails)
	log.Info("------usermails--count-:%d", len(allmails))
	if len(allmails) > mailMaxCount {
		for i := 0; i < len(allmails)-mailMaxCount; i++ {
			mailid := allmails[i]

			log.Info("----id:%d", mailid)
			if mailid <= 0 {

				continue
			}

			//db.DbOne.GetMailRewards(user.Uid, v.Rewards, mailid)
			mail.getMailRewards(uid, mailid)

		}

		newmailstr := ""
		for i := len(allmails) - mailMaxCount; i < len(allmails); i++ {

			newmailstr = newmailstr + "," + strconv.Itoa(allmails[i])

		}
		db.DbOne.SetPlayerOneInfo(uid, "userbaseinfo", "mails_id", newmailstr)

	}

}

//获取未读或未领取邮件数量(新邮件)
func (mail *Mail) getNewMailNum(uid int) int {
	mailnum := 0

	usermails := ""
	err := db.DbOne.GetPlayerOneInfo(uid, "userbaseinfo", "mails_id", &usermails)
	//log.Info("sermails:%s", usermails)
	if err != nil || usermails == "" {
		return mailnum
	}
	//log.Info("fdsfds")

	mails := utils.SplitStringToIntArray(usermails)

	//allmails := strings.Split(usermails, ",")
	if len(mails) > 0 {
		for i := 0; i < len(mails); i++ {
			//			mailid, err := strconv.Atoi(allmails[i])
			//			if err != nil {
			//				continue
			//			}
			mailid := mails[i]
			if mailid < 0 {
				continue
			}
			isget := -1
			rewardstr := ""
			err = db.DbOne.GetPlayerManyInfo(mailid, "mail", "rewardstr,getstate", &rewardstr, &isget)
			log.Info("--isget%d---rewardstr:%s", isget, rewardstr)

			rewards := []conf.RewardsConfig{}
			err = json.Unmarshal([]byte(rewardstr), &rewards)
			if err != nil {
				continue
			}

			if isget == 0 && len(rewards) > 0 {
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
func (mail *Mail) getMailRewards(uid int, mailid int) bool {
	//user.lock.RLock()
	//defer user.lock.RUnlock()
	lockid := uid % 1000
	mail.UserLock[lockid].Lock()
	defer mail.UserLock[lockid].Unlock()

	err := db.DbOne.GetMailRewards(uid, mailid)

	if err == nil {
		return true
	}
	return false

}

//获取邮件信息
func (mail *Mail) getMailInfo(uid int, count int) *datamsg.SC_MailInfo {

	jd := &datamsg.SC_MailInfo{}
	if count <= 0 {
		return jd
	}

	jd.Mails = make([]datamsg.MailInfo, count)
	usermails := ""
	err := db.DbOne.GetPlayerOneInfo(uid, "userbaseinfo", "mails_id", &usermails)
	if err != nil || usermails == "" {
		return jd
	}
	ids := utils.SplitStringToIntArray(usermails)
	if len(ids) > 0 {
		db.DbOne.GetMailInfo(ids, &jd.Mails)
	}

	return jd

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
