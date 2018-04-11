package datamsg

import (
	"encoding/json"
)

var LoginMode = "Login"   //登录模块
var GateMode = "Gate"     //网关模块
var ClientMode = "Client" //客户端模块

var HallMode = "Hall"     //大厅模块
var Game5GMode = "Game5G" //五指棋游戏模块

//消息类型
var SC_Heart = "SC_Heart"
var CS_Heart = "CS_Heart"

type MsgBase struct {
	ModeType  string
	Uid       int
	MsgId     int
	MsgType   string
	ConnectId int
	JsonData  string
}

//玩家信息
type MsgPlayerInfo struct {
	Uid       int
	Name      string
	Gold      int64
	WinCount  int
	LoseCount int
}

//微信登录
type CS_MsgWeiXingLogin struct {
	Code string
	Name string
}

//快速登录
type CS_MsgQuickLogin struct {
	Platform  string
	MachineId string
	//Abc			[10][10]int
}

//进入游戏
type CS_GoIn struct {
	GameId int
}

//退出游戏
type CS_GoOut struct {
	GameId int
}

//走棋
type CS_DoGame5G struct {
	X int
	Y int
}

//获取当前进行中的游戏信息
type CS_GetGamingInfo struct {
	Count int //数量
}

//当前进行中的游戏信息
type MsgGame5GingInfo struct {
	GameId        int
	GameName      string
	PlayerOneName string
	PlayerTwoName string
	Score         int
}

//当前进行中的游戏信息
type SC_GetGamingInfo struct {
	GameInfo []MsgGame5GingInfo
}

//CS_GoIn

//大厅信息
type SC_MsgHallInfo struct {
	PlayerInfo MsgPlayerInfo
}

//操作结果
type SC_Result struct {
	Result string
}

//游戏创建好了 等到加入
type SC_NewGame struct {
	GameId int
}

//玩家
type MsgGame5GPlayerInfo struct {
	//基本数据
	Uid       int
	Name      string
	Gold      int64
	WinCount  int
	LoseCount int

	//游戏中数据
	SeatIndex  int //座位号
	PlayerType int //玩家类型 //玩家类型 1表示玩家 2表示旁观者
	Time       int //剩余总时间
	EveryTime  int //剩余的每次操作时间
}

//游戏信息
type MsgGame5GInfo struct {

	//游戏中数据
	//游戏ID
	GameId int
	//游戏状态
	State     int
	Time      int //总时间
	EveryTime int //每次操作时间
	//该下棋的人的位置号
	GameSeatIndex int

	//棋盘信息
	QiPan [15][15]int
}

//其他玩家进入
type SC_PlayerGoIn struct {
	PlayerInfo MsgGame5GPlayerInfo
}

//游戏信息
type SC_GameInfo struct {
	PlayerInfo  []MsgGame5GPlayerInfo
	ObserveInfo []MsgGame5GPlayerInfo
	GameInfo    MsgGame5GInfo
}

//游戏开始
type SC_GameStart struct {
	GameSeatIndex int
}

//切换下棋的人
type SC_ChangeGameTurn struct {
	GameSeatIndex int
	Time          int
	EveryTime     int
}

//走棋
type SC_DoGame5G struct {
	GameSeatIndex int
	X             int
	Y             int
	Time          int
}

//游戏结束
type SC_GameOver struct {
	WinPlayerSeatIndex int
}

//玩家离开
type SC_PlayerGoOut struct {
	Uid int
}

type MsgRegisterToGate struct {
	ModeType string
}

func NewMsgSC_Result(uid int, connectid int, result string) []byte {
	data := &MsgBase{}
	data.ModeType = "Client"
	data.MsgType = "SC_Result"
	data.Uid = uid
	data.ConnectId = connectid
	jd := &SC_Result{}
	jd.Result = result

	return NewMsg1Bytes(data, jd)
}

func NewMsg1Bytes(data *MsgBase, jd interface{}) []byte {
	if jd != nil {
		jdbytes, _ := json.Marshal(jd)
		data.JsonData = string(jdbytes)
	}
	data1, err1 := json.Marshal(data)
	if err1 == nil {
		return data1
	}
	return []byte("")
}
