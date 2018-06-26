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

	a.Mydb.SetMaxOpenConns(10000)
	a.Mydb.SetMaxIdleConns(500)
	a.Mydb.Ping()
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
		name = "yk_" + strconv.Itoa(int(id))
	}

	isandroid := 0
	if platfom == "android" {
		isandroid = 1
	}

	res, err1 = tx.Exec("INSERT userbaseinfo (uid,name,gold,wincount,losecount,level,experience,seasonscore,avatarurl,firstqizi,secondqizi,isandroid) values (?,?,?,?,?,?,?,?,?,?,?,?)",
		id, name, 0, 0, 0, 1, 0, 1000, "", 1001, 1002, isandroid)
	//插入名字失败
	if err1 != nil {

		name = "yk_" + strconv.Itoa(int(id))
		res, err1 = tx.Exec("INSERT userbaseinfo (uid,name,gold,wincount,losecount,level,experience,seasonscore,avatarurl,firstqizi,secondqizi,isandroid) values (?,?,?,?,?,?,?,?,?,?,?,?)",
			id, name, 0, 0, 0, 1, 0, 1000, "", 1001, 1002, isandroid)
		if err1 != nil {
			log.Info("INSERT userbaseinfo err")
			return -1, tx.Rollback()
		}
	}
	n, e = res.RowsAffected()
	if n == 0 || e != nil {
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
	stmt, err := a.Mydb.Prepare("SELECT name,gold,wincount,losecount,seasonscore,avatarurl,firstqizi,secondqizi,isandroid,RankNum FROM userbaseinfo where uid=?")

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
		return rows.Scan(&info.Name, &info.Gold, &info.WinCount, &info.LoseCount, &info.SeasonScore, &info.AvatarUrl, &info.FirstQiZi, &info.SecondQiZi, &info.IsAndroid, &info.RankNum)
	} else {
		log.Info("no user:%d", uid)
		return errors.New("no user")
	}

}

//获取信息
func (a *DB) GetPlayerManyInfo(uid int, tablename string, field string, value ...interface{}) error {

	if len(field) <= 0 {
		return errors.New("no field")
	}

	idname := "uid"
	if tablename == "mail" || tablename == "publicmail" {
		idname = "id"
	}
	//	fieldstr := field[0]
	//	for i := 1; i < len(field); i++ {
	//		fieldstr = fieldstr + "," + field[i]
	//	}
	sqlstr := "SELECT " + field + " FROM " + tablename + " where " + idname + "=?"

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
		return rows.Scan(value...)
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
		log.Info("update err--%d", n)
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

//重置所有玩家赛季分
func (a *DB) ResetAllPlayerSeasonScore() error {

	sqlstr := "UPDATE userbaseinfo SET seasonscore=ceil(sqrt(abs((seasonscore-1000+abs(seasonscore-1000))/2+1000)/1000)*1000)"

	tx, _ := a.Mydb.Begin()

	res, err1 := tx.Exec(sqlstr)
	//res.LastInsertId()
	n, e := res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("update err")
		return tx.Rollback()
	}

	err1 = tx.Commit()
	return err1
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
	//log.Info(string(jsonData))
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

func (a *DB) UpdateRecord(winuid int, loseuid int) error {
	tx, _ := a.Mydb.Begin()

	twouidstr := utils.GetMinMaxUidStr(winuid, loseuid)
	minuid := utils.GetMin(winuid, loseuid)
	maxuid := utils.GetMax(winuid, loseuid)

	minadd := 1
	maxadd := 1
	if minuid == loseuid {
		minadd = 0
	}
	if maxuid == loseuid {
		maxadd = 0
	}
	log.Info("----UpdateRecord--")

	//基础表
	res, err1 := tx.Exec("UPDATE relation SET win1=win1+?,win2=win2+? where twouid=?", minadd, maxadd, twouidstr)
	//
	if err1 != nil {
		log.Info("UPDATE UpdateRecord err")
		return tx.Rollback()
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		res, err1 = tx.Exec("INSERT relation (twouid,uid1,uid2,win1,win2,u1rewardtime,u2rewardtime) values (?,?,?,?,?,?,?)",
			twouidstr, minuid, maxuid, minadd, maxadd, "2018-06-08 08:50:00", "2018-06-08 08:50:00")
		if err1 != nil {
			log.Error("INSERT relation err %s", err1.Error())
			return tx.Rollback()
		}
	}

	err1 = tx.Commit()
	return err1
}

//相互加为好友
func (a *DB) AddToFriends(uid1 int, uid2 int) error {
	tx, _ := a.Mydb.Begin()

	//基础表
	res, err1 := tx.Exec("UPDATE userbaseinfo SET friends_id=concat_ws(',',friends_id,?) where uid=?", uid1, uid2)
	n, e := res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("UPDATE AddToFriends err")
		return tx.Rollback()
	}
	res, err1 = tx.Exec("UPDATE userbaseinfo SET friends_id=concat_ws(',',friends_id,?) where uid=?", uid2, uid1)
	n, e = res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("UPDATE AddToFriends err")
		return tx.Rollback()
	}

	twouidstr := utils.GetMinMaxUidStr(uid1, uid2)
	minuid := utils.GetMin(uid1, uid2)
	maxuid := utils.GetMax(uid1, uid2)

	//关系表
	res, err1 = tx.Exec("INSERT relation (twouid,uid1,uid2,win1,win2,u1rewardtime,u2rewardtime) values (?,?,?,?,?,?,?)",
		twouidstr, minuid, maxuid, 0, 0, "2018-06-08 08:50:00", "2018-06-08 08:50:00")
	if err1 != nil {
		log.Error("INSERT relation err %s", err1.Error())
		return tx.Rollback()
	}

	err1 = tx.Commit()
	return err1

}

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
	//log.Info("add mail id:%d", int(id))

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

//获取排行榜数据库信息
func (a *DB) GetRankInfo(rank *[]datamsg.RankNodeInfo, seasonidindex int) error {

	sqlstr := "SELECT * FROM rank where seasonidindex=" + strconv.Itoa(seasonidindex)
	str, err := a.GetJSON(sqlstr)
	if err != nil {
		log.Info(err.Error())
		return err
	}

	//h2 := datamsg.MailInfo{}
	err = json.Unmarshal([]byte(str), rank)
	if err != nil {
		log.Info(err.Error())
		return err
	}

	return nil
}

//获取赛季排名
func (a *DB) GetRankNum(seasonidindex int, uid int) int {

	ranknum := 1001
	stmt, err := a.Mydb.Prepare("SELECT ranknum FROM rank where BINARY (seasonidindex=? and uid=?)")

	if err != nil {
		log.Info(err.Error())
		return -1
	}
	defer stmt.Close()
	rows, err := stmt.Query(seasonidindex, uid)
	if err != nil {
		log.Info(err.Error())
		return ranknum
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&ranknum)
	} else {
		//log.Info("no user:%s,%s", machineid, platfom)
	}

	return ranknum
}

//持久化排行榜数据库信息
func (a *DB) WriteRankInfo(rank []datamsg.RankNodeInfo, idindex int) error {
	tx, _ := a.Mydb.Begin()

	_, err := tx.Exec("delete from rank where seasonidindex=" + strconv.Itoa(idindex))
	if err != nil {
		log.Error("delete rank err%s", err.Error())

	}

	for k, v := range rank {
		_, err1 := tx.Exec("INSERT rank (uid,score,name,avatar,rewardgold,seasonidindex,ranknum,AllRankNum) values (?,?,?,?,?,?,?,?)",
			v.Uid, v.Score, v.Name, v.Avatar, v.Rewardgold, idindex, k+1, v.AllRankNum)

		if err1 != nil {
			log.Error("INSERT rank err%s", err1.Error())
			//return tx.Rollback()
		}
	}
	err2 := tx.Commit()
	return err2
}

//获取信息
func (a *DB) GetBattleManyInfo(idstr string, field string, value ...interface{}) error {

	if len(field) <= 0 {
		return errors.New("no field")
	}

	idname := "twouid"

	sqlstr := "SELECT " + field + " FROM relation where " + idname + "=?"

	stmt, err := a.Mydb.Prepare(sqlstr)

	if err != nil {
		log.Info(err.Error())
		return err
	}
	defer stmt.Close()
	rows, err := stmt.Query(idstr)
	if err != nil {
		log.Info(err.Error())
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(value...)
	} else {
		log.Info("no user:%s", idstr)
		return errors.New("no user")
	}

}

//获取好友对战数据库信息
func (a *DB) GetFriendsBattleInfo(myuid int, friendinfo *[]datamsg.FriendInfo) error {

	if friendinfo == nil || len(*friendinfo) <= 0 || myuid <= 0 {
		return nil
	}
	//FriendWin   int //朋友赢得次数
	//	MyWin       int //我赢得次数
	//tx, _ := a.Mydb.Begin()

	for k, v := range *friendinfo {
		idstr := utils.GetMinMaxUidStr(myuid, v.Uid)
		minwin := 0
		maxwin := 0
		a.GetBattleManyInfo(idstr, "win1,win2", &minwin, &maxwin)
		if myuid < v.Uid {
			(*friendinfo)[k].MyWin = minwin
			(*friendinfo)[k].FriendWin = maxwin
		} else {
			(*friendinfo)[k].MyWin = maxwin
			(*friendinfo)[k].FriendWin = minwin
		}
	}

	return nil
}

//获取好友基本数据库信息
func (a *DB) GetFriendsBaseInfo(frienduid []int, friendinfo *[]datamsg.FriendInfo) error {

	if len(frienduid) <= 0 {
		return nil
	}

	sqlstr := "SELECT uid,name,avatarurl,seasonscore,RankNum FROM userbaseinfo where uid = " + strconv.Itoa(frienduid[0])
	for i := 1; i < len(frienduid); i++ {
		sqlstr = sqlstr + " or uid = " + strconv.Itoa(frienduid[i])
	}

	str, err := a.GetJSON(sqlstr)
	if err != nil {
		log.Info(err.Error())
		return err
	}

	//h2 := datamsg.MailInfo{}
	err = json.Unmarshal([]byte(str), friendinfo)
	if err != nil {
		log.Info(err.Error())
		return err
	}

	return nil
}

//获取邮件数据库信息
func (a *DB) GetMailInfo(mailid []int, mail *[]datamsg.MailInfo) error {

	if len(mailid) <= 0 {
		return nil
	}

	sqlstr := "SELECT * FROM mail where id = " + strconv.Itoa(mailid[0])
	for i := 1; i < len(mailid); i++ {
		sqlstr = sqlstr + " or id = " + strconv.Itoa(mailid[i])
	}

	str, err := a.GetJSON(sqlstr)
	if err != nil {
		log.Info(err.Error())
		return err
	}

	//h2 := datamsg.MailInfo{}
	err = json.Unmarshal([]byte(str), mail)
	if err != nil {
		log.Info(err.Error())
		return err
	}
	for k, v := range *mail {

		if len(v.Rewardstr) > 0 {
			err = json.Unmarshal([]byte(v.Rewardstr), &((*mail)[k].Reward))
			if err != nil {
				log.Info(err.Error())
			}
		}
		log.Info("--GetMailInfo--id:%d-----v:%v", v.Id, v)
	}

	return nil
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
		log.Info("err isget%d", isget)
		return err
	}
	recuid := 0
	err = a.GetPlayerOneInfo(mailid, "mail", "recUid", &recuid)
	if err != nil || recuid != uid {
		log.Info("err recuid:%d", recuid)
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
		if n != 0 {
			log.Info("n:%d", n)
		}

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
		//log.Info("---len:%d", len(values))
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

//增加玩家游戏结束后的 奖励分数
func (a *DB) AddPlayerScore(uid int, score int) error {

	if score <= 0 || uid <= 0 {
		return nil
	}

	tx, _ := a.Mydb.Begin()

	res, err1 := tx.Exec("UPDATE userbaseinfo SET seasonscore=seasonscore+?,gameover_addscore=? where uid=? and gameover_addscore<=0", score, score, uid)
	//res.LastInsertId()
	n, e := res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("update err")
		return tx.Rollback()
	}

	err1 = tx.Commit()
	return err1

}

//更新玩家胜负 事务
func (a *DB) UpdatePlayerWinLose(winid int, winseasonscore int, loseid int, loseseasonscore int) error {
	tx, _ := a.Mydb.Begin()

	if winseasonscore > 0 {
		res, err1 := tx.Exec("UPDATE userbaseinfo SET wincount=wincount+1,seasonscore=seasonscore+?,gameover_addscore=0 where uid=?", winseasonscore, winid)
		//res.LastInsertId()
		n, e := res.RowsAffected()
		if err1 != nil || n == 0 || e != nil {
			log.Info("update err")
			return tx.Rollback()
		}
	} else {
		res, err1 := tx.Exec("UPDATE userbaseinfo SET wincount=wincount+1,seasonscore=seasonscore+? where uid=?", winseasonscore, winid)
		//res.LastInsertId()
		n, e := res.RowsAffected()
		if err1 != nil || n == 0 || e != nil {
			log.Info("update err")
			return tx.Rollback()
		}
	}

	if loseseasonscore > 0 {
		//重置gameover_addscore 游戏结束后的加分值
		res, err1 := tx.Exec("UPDATE userbaseinfo SET losecount=losecount+1,seasonscore=seasonscore-?,gameover_addscore=0 where uid=?", loseseasonscore, loseid)
		n, e := res.RowsAffected()
		if err1 != nil || n == 0 || e != nil {
			log.Info("update err")
			return tx.Rollback()
		}
	} else {
		res, err1 := tx.Exec("UPDATE userbaseinfo SET losecount=losecount+1,seasonscore=seasonscore-? where uid=?", loseseasonscore, loseid)
		n, e := res.RowsAffected()
		if err1 != nil || n == 0 || e != nil {
			log.Info("update err")
			return tx.Rollback()
		}
	}

	err1 := tx.Commit()
	return err1

}

//更新玩家头像
func (a *DB) UpdatePlayerAvatarAndName(avatarurl string, name string, uid int) error {
	tx, _ := a.Mydb.Begin()
	//	res := nil
	//	err1 := nil

	//if len(name) > 0 {
	res, err1 := tx.Exec("UPDATE userbaseinfo SET avatarurl=?,name=? where uid=?", avatarurl, name, uid)
	//	} else {
	//		res, err1 := tx.Exec("UPDATE userbaseinfo SET avatarurl=? where uid=?", avatarurl, uid)
	//	}

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
