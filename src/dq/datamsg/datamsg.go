package datamsg

import (
	"dq/conf"
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
	Uid         int
	Name        string
	Gold        int64
	WinCount    int
	LoseCount   int
	SeasonScore int
	AvatarUrl   string
	FirstQiZi   int
	SecondQiZi  int
	IsAndroid   int
}

//微信登录
type CS_MsgWeiXingLogin struct {
	Code      string
	Name      string
	AvatarUrl string //头像
}

//快速登录
type CS_MsgQuickLogin struct {
	Platform  string
	MachineId string
	//Abc			[10][10]int
}

//进入游戏
type CS_GoIn struct {
	GameId         int
	OtherPlayerUid int
}

//退出游戏
type CS_GoOut struct {
	GameId int
}

//走棋11
type CS_DoGame5G struct {
	X int
	Y int
}

//获取当前进行中的游戏信息
type CS_GetGamingInfo struct {
	Count int //数量
}

//查看能否进入此房间
type CS_CheckGoToGame struct {
	GameId         int
	CreateGameTime int64
}

//获取任务奖励
type CS_GetTaskRewards struct {
	Id int //任务id
}

//获取邮件奖励
type CS_GetMailRewards struct {
	Id int //任务id
}

//购买商品
type CS_BuyItem struct {
	Id    int //任务id
	Index int
}

//购买商品
type CS_ZhuangBeiItem struct {
	Type int
}

//邀请好友来战
type CS_YaoQingFriend struct {
	MyName    string
	FriendUid int
	GameId    int
}

//邀请好友来战
type SC_YaoQingFriend struct {
	Name   string
	Uid    int
	GameId int
}

//获取排行信息
type CS_GetRankInfo struct {
	StartRank int
	EndRank   int
}

//上传推荐者
type CS_Presenter struct {
	PresenterUid int
}

//大厅信息
type SC_GetTaskRewards struct {
	Code int //1表示成功
	Id   int //任务ID
}

//大厅信息
type SC_GetMailRewards struct {
	Code int //1表示成功
	Id   int //任务ID
}

//大厅信息
type SC_CheckGoToGame struct {
	GameId int
	Code   int    //1表示可以进入 0表示不能进入
	Err    string //不能进入原因
}

//当前进行中的游戏信息
type MsgGame5GingInfo struct {
	GameId        int
	GameName      string
	PlayerOneName string
	PlayerTwoName string
	ScoreOne      int
	ScoreTwo      int
	AvatarOne     string
	AvatarTwo     string
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

//大厅UI信息
type SC_HallUIInfo struct {
	TaskED_ShowNum int
	Task_ShowNum   int
	Mail_ShowNum   int
}

//操作结果
type SC_Result struct {
	Result string
}

//游戏创建好了 等到加入
type SC_NewGame struct {
	GameId int
}

//GameOverInfo
//游戏结束信息
type GameOverInfo struct {
	GameMode        int
	WinId           int
	LoseId          int
	ObserverId      []int
	WinPlayerScore  int
	LosePlayerScore int
}

//
type GameStateChangeInfo struct {
	Uid    int
	State  int
	GameId int
}

//玩家
type MsgGame5GPlayerInfo struct {
	//基本数据
	Uid             int
	Name            string
	Gold            int64
	WinCount        int
	LoseCount       int
	SeasonScore     int
	AvatarUrl       string
	QiZiId          int
	Qizi_move       int
	Qizi_move_trail int
	Qizi_floor      int
	Qizi_lastplay   int
	Beiyongtime     int
	Steptime        int

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
	//游戏模式
	GameMode int

	//棋盘信息
	QiPan [15][15]int

	//游戏创建时间戳
	CreateGameTime int64
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
	GameSeatIndex     int
	SeatIndex0_qiziid int
	SeatIndex1_qiziid int
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
	Reason             int //0表示玩家退出 1表示时间到 2表示棋子5连
	WinQiZi            [5][2]int
	WinScore           int
	LoseScore          int
}

type MsgTaskInfo struct {
	Id            int
	Type          int
	DestValue     int //目标
	ProgressValue int //进度
	State         int //状态 0表示未完成 1表示已完成没领取 2表示已领取

	Rewards []conf.RewardsConfig
}
type RankNodeInfo struct {
	Uid        int
	Score      int
	Name       string
	Avatar     string
	Rewardgold int
}

type RankNodeMessage struct {
	RankNodeInfo
	RankNum int
}

type MailInfo struct {
	Id        int
	SendName  string
	Title     string
	Content   string
	RecUid    int
	Date      string
	Reward    []conf.RewardsConfig
	Rewardstr string
	ReadState int
	GetState  int
}

type FriendInfo struct {
	Uid         int
	Name        string
	Avatar      string
	Seasonscore int
	State       int //0离线 1大厅 2比赛中 3观战中
	FriendWin   int //朋友赢得次数
	MyWin       int //我赢得次数
	GiveState   int //打赏状态 0表示还没打赏 1表示已经打赏

}

//好友信息
type SC_FriendsInfo struct {
	Friends []FriendInfo
}

type SC_RankInfo struct {
	Ranks      []RankNodeMessage
	SeasonInfo conf.Season
	MyRank     int
}

//邮件信息
type SC_MailInfo struct {
	Mails []MailInfo
}

type StoreInfo struct {
	conf.Commodity
	IsStartSale bool
}

type ItemInfo struct {
	Type       int
	ExpiryTime string //到期时间
	IsExpiry   bool   //是否到期
}

//商店信息
type SC_StoreInfo struct {
	Commoditys []StoreInfo
}

//	firstqizi := 0
//	secondqizi := 0
//	qizi_move := 0
//	qizi_move_trail := 0
//	qizi_floor := 0
//	qizi_lastplay := 0
//道具信息
type SC_ItemInfo struct {
	Firstqizi       int
	Secondqizi      int
	Qizi_move       int
	Qizi_move_trail int
	Qizi_floor      int
	Qizi_lastplay   int
	Beiyongtime     int
	Steptime        int
	Items           []ItemInfo
}
type SC_UpdateUsedItem struct {
	Firstqizi       int
	Secondqizi      int
	Qizi_move       int
	Qizi_move_trail int
	Qizi_floor      int
	Qizi_lastplay   int
	Beiyongtime     int
	Steptime        int
}

//每日任务信息
type SC_TskEdInfo struct {
	Task []MsgTaskInfo
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
