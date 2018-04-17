package db

import (
	"database/sql"
	"dq/conf"
	"dq/datamsg"
	"dq/log"
	"errors"
	"strconv"

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

	id := a.newUser("", "", "", openid)
	if id > 0 {
		if err := a.newUserBaseInfo(id, name); err == nil {
			return id
		}
	}
	return id
}

//创建快速新玩家
func (a *DB) CreateQuickPlayer(machineid string, platfom string, name string) int {

	id := a.newUser(machineid, platfom, "", "")
	if id > 0 {
		if err := a.newUserBaseInfo(id, name+strconv.Itoa(id)); err == nil {
			return id
		}
	}
	return id
}

//创建新玩家基础信息
func (a *DB) newUserBaseInfo(id int, name string) error {

	stmt, err := a.Mydb.Prepare(`INSERT userbaseinfo (uid,name,gold,wincount,losecount,level,experience,seasonscore) values (?,?,?,?,?,?,?,?)`)
	defer stmt.Close()
	if err != nil {
		log.Info(err.Error())
		return err
	}
	_, err1 := stmt.Exec(id, name, 0, 0, 0, 1, 0, 1000)
	return err1
}

//创建新玩家信息
func (a *DB) newUser(machineid string, platfom string, phonenumber string, openid string) int {

	stmt, err := a.Mydb.Prepare(`INSERT user (phonenumber,platform,machineid,wechat_id) values (?,?,?,?)`)
	defer stmt.Close()
	if err != nil {
		log.Info(err.Error())
		return -1
	}
	res, err1 := stmt.Exec(phonenumber, platfom, machineid, openid)
	if err1 == nil {
		id, _ := res.LastInsertId()
		return int(id)
	}
	log.Info(err1.Error())
	stmt.Query()
	return -1
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
	stmt, err := a.Mydb.Prepare("SELECT name,gold,wincount,losecount,seasonscore FROM userbaseinfo where uid=?")

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
		return rows.Scan(&info.Name, &info.Gold, &info.WinCount, &info.LoseCount, &info.SeasonScore)
	} else {
		log.Info("no user:%d", uid)
		return errors.New("no user")
	}

}

//更新玩家胜负
func (a *DB) UpdatePlayerWinLose(winid int, winseasonscore int, loseid int, loseseasonscore int) error {
	tx, _ := a.Mydb.Begin()

	res, err1 := tx.Exec("UPDATE userbaseinfo SET wincount=wincount+1,seasonscore=seasonscore+? where uid=?", winseasonscore, winid)
	//res.LastInsertId()
	n, e := res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("update err")
		return tx.Rollback()
	}
	//	res, err1 = tx.Exec("UPDATE userbaseinfo SET seasonscore=seasonscore+? where uid=?", winseasonscore, loseid)
	//	n, e = res.RowsAffected()
	//	if err1 != nil || n == 0 || e != nil {
	//		log.Info("update err")
	//		return tx.Rollback()
	//	}

	res, err1 = tx.Exec("UPDATE userbaseinfo SET losecount=losecount+1,seasonscore=seasonscore-? where uid=?", loseseasonscore, loseid)
	n, e = res.RowsAffected()
	if err1 != nil || n == 0 || e != nil {
		log.Info("update err")
		return tx.Rollback()
	}

	//	res, err1 = tx.Exec("UPDATE userbaseinfo SET seasonscore=seasonscore-? where uid=?", loseseasonscore, loseid)
	//	n, e = res.RowsAffected()
	//	if err1 != nil || n == 0 || e != nil {
	//		log.Info("update err")
	//		return tx.Rollback()
	//	}
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
