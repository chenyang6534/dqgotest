package hall

import (
	//"dq/conf"
	//"dq/datamsg"
	"dq/db"
	//"dq/log"
	"dq/utils"
	"sync"
	//"time"
)

var (
	mail = &Mail{Lock: new(sync.Mutex)}
)

type Mail struct {
	PublicMail *utils.BeeMap
	MaxIndex   int
	Lock       *sync.Mutex
}

func GetMail() *Mail {
	return mail
}

func (mail *Mail) CheckUserPublicMail(uid int, publicIndex int) {
	mail.Lock.Lock()
	defer mail.Lock.Unlock()

	if mail.MaxIndex > publicIndex {

	}
}

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

func (mail *Mail) Init() {
	mail.Lock.Lock()
	defer mail.Lock.Unlock()

	mail.PublicMail = utils.NewBeeMap()
	mail.MaxIndex = 0

	db.DbOne.GetPublicMailInfo(mail.PublicMailm, &mail.MaxIndex)
}
