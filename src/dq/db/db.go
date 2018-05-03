package db

import (
	"database/sql"
	"dq/conf"
	"dq/datamsg"
	"dq/log"
	"dq/utils"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DB struct {
	Mydb *sql.DB
}

var DbOne *DB

func CreateDB() {
	DbOne = new(DB)
	DbOne.Init()
}

func (a *DB) Init() {

	ip := conf.Conf.DataBaseInfo["Ip"].(string)
	nameandpassword := conf.Conf.DataBaseInfo["NameAndPassword"].(string)
	databasename := conf.Conf.DataBaseInfo["DataBaseName"].(string)
	db, err := sql.Open("mysql", nameandpassword+"@"+ip+"/"+databasename)
	if err != nil {
		log.Error(err.Error())
	}
	err = db.Ping()
	if err != nil {
		log.Error(err.Error())
	}
	a.Mydb = db

}

//创建微信openid玩家
func (a *DB) CreateQuickWSOpenidPlayer(openid string, name string) int {

	id, _ := a.newUser("", "", "", openid, name)
	//	if id > 0 {
	//		if err := a.newUserBaseInfo(id, name); err == nil {
	//			return err
	//		}

	//	}
	return id
}

//创建快速新玩家
func (a *DB) CreateQuickPlayer(machineid string, platfom string, name string) int {

	id, _ := a.newUser(machineid, platfom, "", "", name)
	//	if id > 0 {
	//		if err := a.newUserBaseInfo(id, name+strconv.Itoa(id)); err == nil {
	//			return err
	//		}

	//	}
	return id
}

//创建新玩家基础信息
func (a *DB) newUserBaseInfo(id int, name string) error {

	//	stmt, err := a.Mydb.Prepare(`INSERT userbaseinfo (uid,name,gold,wincount,losecount,level,experience,seasonscore,avatarurl,firstqizi,secondqizi) values (?,?,?,?,?,?,?,?,?,?,?)`)
	//	defer stmt.Close()
	//	if err != nil {
	//		log.Info(err.Error())
	//		return err
	//	}
	//	_, err1 := stmt.Exec(id, name, 0, 0, 0, 1, 0, 1000, "", 1, 2)
	//	return err1

	tx, _ := a.Mydb.Begin()

	res, err1 := tx.Exec("INSERT userbaseinfo (uid,name,gold,wincount,losecount,level,experience,seasonscore,avatarurl,firstqizi,secondqizi) values (?,?,?,?,?,?,?,?,?,?,?)",
		id, name, 0, 0, 0, 1, 0, 1000, "", 1, 2)
	n, e := res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("INSERT userbaseinfo err")
		return tx.Rollback()
	}

	day := time.Now().Format("2006-01-02")
	res, err1 = tx.Exec("INSERT taskeveryday (uid,day) values (?,?)", id, day)
	n, e = res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("INSERT taskeveryday err")
		return tx.Rollback()
	}

	err1 = tx.Commit()
	return err1

}

////创建新玩家每日任务信息
//func (a *DB) newUserTaskEverydayInfo(id int) error {

//	day := time.Now().Format("2006-01-02")

//	stmt, err := a.Mydb.Prepare(`INSERT taskeveryday (uid,day) values (?,?)`)
//	defer stmt.Close()
//	if err != nil {
//		log.Info(err.Error())
//		return err
//	}
//	_, err1 := stmt.Exec(id, day)
//	return err1
//}

//创建新玩家信息
func (a *DB) newUser(machineid string, platfom string, phonenumber string, openid string, name string) (int, error) {

	tx, _ := a.Mydb.Begin()

	res, err1 := tx.Exec("INSERT user (phonenumber,platform,machineid,wechat_id) values (?,?,?,?)",
		phonenumber, platfom, machineid, openid)
	n, e := res.RowsAffected()
	id, err2 := res.LastInsertId()
	if err1 != nil || n == 0 || e != nil || err2 != nil {
		log.Info("INSERT user err")
		return -1, tx.Rollback()
	}
	if name == "" {
		name = "xiaoming_" + strconv.Itoa(int(id))
	}

	res, err1 = tx.Exec("INSERT userbaseinfo (uid,name,gold,wincount,losecount,level,experience,seasonscore,avatarurl,firstqizi,secondqizi) values (?,?,?,?,?,?,?,?,?,?,?)",
		id, name, 0, 0, 0, 1, 0, 1000, "", 1, 2)
	n, e = res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("INSERT userbaseinfo err")
		return -1, tx.Rollback()
	}

	day := time.Now().Format("2006-01-02")
	res, err1 = tx.Exec("INSERT taskeveryday (uid,day) values (?,?)", id, day)
	n, e = res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("INSERT taskeveryday err")
		return -1, tx.Rollback()
	}

	err1 = tx.Commit()
	if err1 == nil {
		return int(id), nil
	}
	return -1, err1

	//	stmt, err := a.Mydb.Prepare(`INSERT user (phonenumber,platform,machineid,wechat_id) values (?,?,?,?)`)
	//	defer stmt.Close()
	//	if err != nil {
	//		log.Info(err.Error())
	//		return -1
	//	}
	//	res, err1 := stmt.Exec(phonenumber, platfom, machineid, openid)
	//	if err1 == nil {
	//		id, _ := res.LastInsertId()
	//		return int(id)
	//	}
	//	log.Info(err1.Error())
	//	stmt.Query()
	//	return -1
}

//检查微信openid登录
func (a *DB) CheckWSOpenidLogin(openid string) int {
	var uid int

	stmt, err := a.Mydb.Prepare("SELECT uid FROM user where BINARY wechat_id=? ")

	if err != nil {
		log.Info(err.Error())
		return -1
	}
	defer stmt.Close()
	rows, err := stmt.Query(openid)
	if err != nil {
		log.Info(err.Error())
		return uid
		//创建新账号
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&uid)
	} else {
		log.Info("no user:%s", openid)
	}

	return uid

}

//检查快速登录
func (a *DB) CheckQuickLogin(machineid string, platfom string) int {
	var uid int

	stmt, err := a.Mydb.Prepare("SELECT uid FROM user where BINARY (machineid=? and platform=?)")

	if err != nil {
		log.Info(err.Error())
		return -1
	}
	defer stmt.Close()
	rows, err := stmt.Query(machineid, platfom)
	if err != nil {
		log.Info(err.Error())
		return uid
		//创建新账号
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&uid)
	} else {
		log.Info("no user:%s,%s", machineid, platfom)
	}

	return uid

}

//获取玩家基本信息
func (a *DB) GetPlayerInfo(uid int, info *datamsg.MsgPlayerInfo) error {
	info.Uid = uid
	stmt, err := a.Mydb.Prepare("SELECT name,gold,wincount,losecount,seasonscore,avatarurl,firstqizi,secondqizi FROM userbaseinfo where uid=?")

	if err != nil {
		log.Info(err.Error())
		return err
	}
	defer stmt.Close()
	rows, err := stmt.Query(uid)
	if err != nil {
		log.Info(err.Error())
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&info.Name, &info.Gold, &info.WinCount, &info.LoseCount, &info.SeasonScore, &info.AvatarUrl, &info.FirstQiZi, &info.SecondQiZi)
	} else {
		log.Info("no user:%d", uid)
		return errors.New("no user")
	}

}

//获取信息
func (a *DB) GetPlayerOneInfo(uid int, tablename string, field string, value interface{}) error {

	sqlstr := "SELECT " + field + " FROM " + tablename + " where uid=?"
	if tablename == "mail" || tablename == "publicmail" {
		sqlstr = "SELECT " + field + " FROM " + tablename + " where id=?"
	}

	stmt, err := a.Mydb.Prepare(sqlstr)

	if err != nil {
		log.Info(err.Error())
		return err
	}
	defer stmt.Close()
	rows, err := stmt.Query(uid)
	if err != nil {
		log.Info(err.Error())
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(value)
	} else {
		log.Info("no user:%d", uid)
		return errors.New("no user")
	}

}

//设置信息
func (a *DB) SetPlayerOneInfo(uid int, tablename string, field string, value interface{}) error {

	sqlstr := "UPDATE " + tablename + " SET " + field + "=? where uid=?"
	if tablename == "mail" || tablename == "publicmail" {
		sqlstr = "UPDATE " + tablename + " SET " + field + "=? where id=?"
	}
	tx, _ := a.Mydb.Begin()

	res, err1 := tx.Exec(sqlstr, value, uid)
	//res.LastInsertId()
	n, e := res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("update err")
		return tx.Rollback()
	}

	err1 = tx.Commit()
	return err1

}

//获取玩家一项其他信息
func (a *DB) GetPlayerOneOtherInfo(uid int, field string, value interface{}) error {

	sqlstr := "SELECT " + field + " FROM userbaseinfo where uid=?"

	stmt, err := a.Mydb.Prepare(sqlstr)

	if err != nil {
		log.Info(err.Error())
		return err
	}
	defer stmt.Close()
	rows, err := stmt.Query(uid)
	if err != nil {
		log.Info(err.Error())
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(value)
	} else {
		log.Info("no user:%d", uid)
		return errors.New("no user")
	}

}

//设置玩家一项其他信息
func (a *DB) SetPlayerOneOtherInfo(uid int, field string, value interface{}) error {

	sqlstr := "UPDATE userbaseinfo SET " + field + "=? where uid=?"

	tx, _ := a.Mydb.Begin()

	res, err1 := tx.Exec(sqlstr, value, uid)
	//res.LastInsertId()
	n, e := res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("update err")
		return tx.Rollback()
	}

	err1 = tx.Commit()
	return err1
}

//获取玩家其他信息
func (a *DB) GetPlayerOtherInfo(uid int, mails *string, publicIndex *int) error {

	stmt, err := a.Mydb.Prepare("SELECT mails_id,publicmail_index FROM userbaseinfo where uid=?")

	if err != nil {
		log.Info(err.Error())
		return err
	}
	defer stmt.Close()
	rows, err := stmt.Query(uid)
	if err != nil {
		log.Info(err.Error())
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(mails, publicIndex)
	} else {
		log.Info("no user:%d", uid)
		return errors.New("no user")
	}

}

func (a *DB) GetJSON(sqlString string) (string, error) {
	stmt, err := a.Mydb.Prepare(sqlString)
	if err != nil {
		return "", err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return "", err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}
	count := len(columns)
	tableData := make([]map[string]interface{}, 0)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)
		entry := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		tableData = append(tableData, entry)
	}
	jsonData, err := json.Marshal(tableData)
	if err != nil {
		return "", err
	}
	log.Info(string(jsonData))
	return string(jsonData), nil
}

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
//给用户写邮件
func (a *DB) WritePrivateMailInfo(uid int, mailInfo *datamsg.MailInfo) error {
	tx, _ := a.Mydb.Begin()

	//	rewardlen := len(mailInfo.Reward)
	//	reward := [3]string{"", "", ""}
	//	for i := 0; i < rewardlen; i++ {
	//		tt, _ := json.Marshal(mailInfo.Reward[i])
	//		reward[i] = string(tt)
	//		log.Info("------------i:%d----str:%s", i, reward[i])
	//	}
	rewards, _ := json.Marshal(mailInfo.Reward)

	res, err1 := tx.Exec("INSERT mail (sendname,title,content,recUid,date,rewardstr,readstate,getstate) values (?,?,?,?,?,?,?,?)",
		mailInfo.SendName, mailInfo.Title, mailInfo.Content, uid, mailInfo.Date, rewards, mailInfo.ReadState, mailInfo.GetState)
	n, e := res.RowsAffected()
	id, err2 := res.LastInsertId()
	if err1 != nil || n == 0 || e != nil || err2 != nil {
		log.Info("INSERT mail err")
		return tx.Rollback()
	}

	addmail := strconv.Itoa(int(id))

	res, err1 = tx.Exec("UPDATE userbaseinfo SET mails_id=concat_ws(',',mails_id,?) where uid=?", addmail, uid)
	n, e = res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("UPDATE mail err")
		return tx.Rollback()
	}

	err1 = tx.Commit()
	return err1
}
func (a *DB) WritePrivateMailInfoFromPublic(uid int, mailInfo *datamsg.MailInfo, index int) error {
	tx, _ := a.Mydb.Begin()

	//	rewardlen := len(mailInfo.Reward)
	//	reward := [3]string{"", "", ""}
	//	for i := 0; i < rewardlen; i++ {
	//		tt, _ := json.Marshal(mailInfo.Reward[i])
	//		reward[i] = string(tt)
	//		log.Info("------------i:%d----str:%s", i, reward[i])
	//	}

	rewards, _ := json.Marshal(mailInfo.Reward)

	res, err1 := tx.Exec("INSERT mail (sendname,title,content,recUid,date,rewardstr,readstate,getstate) values (?,?,?,?,?,?,?,?)",
		mailInfo.SendName, mailInfo.Title, mailInfo.Content, uid, mailInfo.Date, rewards, mailInfo.ReadState, mailInfo.GetState)
	n, e := res.RowsAffected()
	id, err2 := res.LastInsertId()
	if err1 != nil || n == 0 || e != nil || err2 != nil {
		log.Info("INSERT mail err")
		return tx.Rollback()
	}

	addmail := strconv.Itoa(int(id))

	res, err1 = tx.Exec("UPDATE userbaseinfo SET mails_id=concat_ws(',',mails_id,?) where uid=?", addmail, uid)
	n, e = res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("UPDATE mail err")
		return tx.Rollback()
	}

	res, err1 = tx.Exec("UPDATE userbaseinfo SET publicmail_index=? where uid=?", index, uid)
	n, e = res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("UPDATE mail err")
		return tx.Rollback()
	}

	err1 = tx.Commit()
	return err1
}

//获取公共邮件数据库信息
func (a *DB) GetPublicMailInfo(mails *utils.BeeMap, maxIndex *int) error {

	//stmt, err := a.Mydb.Prepare("SELECT * FROM publicmail")
	*maxIndex = 0

	str, err := a.GetJSON("SELECT * FROM publicmail")
	if err != nil {
		log.Info(err.Error())
		return err
	}

	h2 := []datamsg.MailInfo{}
	err = json.Unmarshal([]byte(str), &h2)
	if err != nil {
		log.Info(err.Error())
		return err
	}
	for _, v := range h2 {
		d1 := v

		if len(v.Rewardstr) > 0 {
			err = json.Unmarshal([]byte(v.Rewardstr), &d1.Reward)
			if err != nil {
				log.Info(err.Error())
			}
		}

		mails.Set(v.Id, &d1)
		if *maxIndex < v.Id {
			*maxIndex = v.Id
		}
		log.Info("----id:%d-----v:%v", d1.Id, d1)
	}

	return nil
}

//type RewardsConfig struct {
//	Type  int
//	Count int
//	Time  int
//}

//获取奖励
func (a *DB) GetMailRewards(uid int, mailid int) error {

	log.Info("GetMailRewards")
	isget := 0
	err := a.GetPlayerOneInfo(mailid, "mail", "getstate", &isget)
	if err != nil || isget != 0 {
		log.Info(err.Error())
		return err
	}
	recuid := 0
	err = a.GetPlayerOneInfo(mailid, "mail", "recUid", &recuid)
	if err != nil || recuid != uid {
		log.Info(err.Error())
		return err
	}

	//rewards := make([]conf.RewardsConfig, 0)
	//for i := 0; i < 3; i++ {
	reward := ""
	err = a.GetPlayerOneInfo(mailid, "mail", "rewardstr", &reward)
	if err != nil || reward == "" {
		return err
	}
	rewards := []conf.RewardsConfig{}
	err = json.Unmarshal([]byte(reward), &rewards)
	if err != nil {
		log.Info(err.Error())
		return err
	}
	//rewards = append(rewards, rw)

	//}

	tx, _ := a.Mydb.Begin()

	for _, v := range rewards {
		err = a.doRewards(tx, uid, v)
		if err != nil {
			return err
		}
	}

	err1 := tx.Commit()
	if err1 == nil {
		err1 = a.SetPlayerOneInfo(mailid, "mail", "getstate", 1)
	}

	return err1
}

func (a *DB) doRewards(tx *sql.Tx, uid int, reward conf.RewardsConfig) error {
	//金币
	v := reward
	if v.Type == 1 {
		res, err1 := tx.Exec("UPDATE userbaseinfo SET gold=gold+? where uid=?", v.Count, uid)
		//res.LastInsertId()
		n, e := res.RowsAffected()
		if err1 != nil || n == 0 || e != nil {
			log.Info("update err")
			return tx.Rollback()
		}
	}

	return nil

}

//获取奖励
func (a *DB) GetTaskRewards(uid int, rewards []conf.RewardsConfig, getTagDBFieldName string) error {

	tx, _ := a.Mydb.Begin()

	for _, v := range rewards {
		err := a.doRewards(tx, uid, v)
		if err != nil {
			return err
		}
	}

	sqlstr := "UPDATE taskeveryday SET " + getTagDBFieldName + "=1 where uid=?"
	res, err1 := tx.Exec(sqlstr, uid)
	//res.LastInsertId()
	n, e := res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("update err")
		return tx.Rollback()
	}

	err1 = tx.Commit()
	return err1
}

//获取玩家每日任务信息
func (a *DB) SetPlayerTaskEd(uid int, date string, info *utils.BeeMap) error {

	//----------
	size := 1
	if info != nil {
		size += info.Size()
	}
	keys := make([]string, size)
	values := make([]interface{}, size)

	sqlstr := "UPDATE taskeveryday SET "

	count := 0
	keys[count] = "day"
	values[count] = "\"" + date + "\""
	//sqlstr = sqlstr + keys[count] + "=" + (values[count].(string))
	sqlstr = sqlstr + keys[count] + "=?"
	count++

	if info != nil {
		items := info.Items()
		for k, v := range items {
			keys[count] = k.(string)
			values[count] = v
			count++
		}
	}

	for i := 1; i < size; i++ {
		sqlstr = sqlstr + "," + keys[i] + "=" + strconv.Itoa(values[i].(int))
	}

	sqlstr += " where uid=?"

	tx, _ := a.Mydb.Begin()

	res, err1 := tx.Exec(sqlstr, date, uid)

	n, e := res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		if err1 != nil {
			log.Error(err1.Error())
		}
		if e != nil {
			log.Error(e.Error())
		}

		log.Info("n:%d", n)
		return tx.Rollback()
	}

	err1 = tx.Commit()
	return err1
}

//获取玩家每日任务信息
func (a *DB) GetPlayerTaskEd(uid int, date *string, info *utils.BeeMap) error {

	size := 1
	if info != nil {
		size += info.Size()
	}
	count := 0
	keys := make([]string, size)
	values := make([]interface{}, size)
	data := make([]int, size)
	for i := 0; i < size; i++ {
		values[i] = &data[i]
	}
	values[0] = date

	sqlstr := "SELECT "
	keys[count] = "day"
	//values[count] = "2016"
	count++
	sqlstr += "day"

	if info != nil {
		items := info.Items()
		for k, _ := range items {
			keys[count] = k.(string)
			//values[count] = v
			count++
			if count == 1000 {
				sqlstr = sqlstr + k.(string)
			} else {
				sqlstr = sqlstr + "," + k.(string)
			}

		}
	}
	sqlstr = sqlstr + " FROM taskeveryday where uid=?"

	//info.Uid = uid

	stmt, err := a.Mydb.Prepare(sqlstr)

	if err != nil {
		log.Info(err.Error())
		return err
	}
	defer stmt.Close()
	rows, err := stmt.Query(uid)
	if err != nil {
		log.Info(err.Error())
		return err
	}
	defer rows.Close()

	if rows.Next() {
		log.Info("---len:%d", len(values))
		//return rows.Scan(&info.Name, &info.Gold, &info.WinCount, &info.LoseCount, &info.SeasonScore, &info.AvatarUrl, &info.FirstQiZi, &info.SecondQiZi)
		err = rows.Scan(values...)
		if err != nil {
			log.Info(err.Error())
			return err
		}
		for i := 1; i < len(keys); i++ {
			info.Set(keys[i], data[i])
		}
		//*date = data[0]
		return nil
	} else {
		log.Info("no user:%d", uid)
		return errors.New("no user")
	}

}

//更新玩家胜负 事务
func (a *DB) UpdatePlayerWinLose(winid int, winseasonscore int, loseid int, loseseasonscore int) error {
	tx, _ := a.Mydb.Begin()

	res, err1 := tx.Exec("UPDATE userbaseinfo SET wincount=wincount+1,seasonscore=seasonscore+? where uid=?", winseasonscore, winid)
	//res.LastInsertId()
	n, e := res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("update err")
		return tx.Rollback()
	}

	res, err1 = tx.Exec("UPDATE userbaseinfo SET losecount=losecount+1,seasonscore=seasonscore-? where uid=?", loseseasonscore, loseid)
	n, e = res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("update err")
		return tx.Rollback()
	}

	err1 = tx.Commit()
	return err1

}

//更新玩家头像
func (a *DB) UpdatePlayerAvatar(avatarurl string, uid int) error {
	tx, _ := a.Mydb.Begin()

	res, err1 := tx.Exec("UPDATE userbaseinfo SET avatarurl=? where uid=?", avatarurl, uid)
	//res.LastInsertId()
	n, e := res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("update err")
		return tx.Rollback()
	}

	err1 = tx.Commit()
	return err1

}

func (a *DB) test() {
	//tx,err := a.Mydb.Begin()
	//tx.Prepare()
}

func (a *DB) Close() {
	a.Mydb.Close()
}
