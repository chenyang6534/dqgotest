package game5g

import(
	"dq/utils"
	"sync"
	"dq/datamsg"
	"dq/timer"
	"time"
	"errors"
	"dq/log"
	"dq/db"
)

//玩家
type Game5GPlayer struct{
	//基本数据
	Uid	int
	ConnectId int
	Name	string
	Gold 	int64
	WinCount int
	LoseCount int
	
	//游戏
	Game 	*Game5GLogic
	
	//游戏中数据
	SeatIndex int //座位号
	PlayerType int //玩家类型 //玩家类型 1表示玩家 2表示旁观者
	Time 		int //剩余总时间
	OperateStartTime time.Duration //操作开始时间
	OperateState int //走棋状态 0 表示待定 1表示走棋中 2表示走棋结束
}

//游戏逻辑

const(
	Game5GState_Wait	= 	1	//等待玩家加入中
	Game5GState_Gaming	= 	2	//游戏中
	Game5GState_Result	= 	3	//结算中
	Game5GState_Over	= 	4	//解散
)




type Game5GLogic struct{
	
	//games
	GameAgent				*Game5GAgent
	
	//游戏ID
	GameId					int
	//将要玩游戏的玩家ID
	WillPlayGamePlayerUid	[2]int
	//玩家
	Player					[2]*Game5GPlayer
	//观看者
	Observer				*utils.BeeMap
	
	//游戏状态
	State					int
	
	//锁
	Lock 					*sync.Mutex
	
	//单人总时间
	Time					int
	//单人单步时间
	EveryTime				int
	
	//该下棋的人的位置号
	GameSeatIndex			int
	
	//棋盘
	QiPan					[15][15]int
	
	//时间到 倒计时
	gameTimer				*timer.Timer
	

}

func (game *Game5GLogic)Init(){
	game.State = Game5GState_Wait
	game.Observer = utils.NewBeeMap()
	game.Lock 	=	new(sync.Mutex)
	
	//初始化棋盘
	for x,value := range game.QiPan {
	    for y,_ := range value {
			game.QiPan[x][y] = -1
	    }
	}
}



func (game *Game5GLogic)notifyAllPlayerGoIn(player *Game5GPlayer){
	
	//
	
	jd := &datamsg.SC_PlayerGoIn{}
	jd.PlayerType = player.PlayerType
	jd.PlayerSeatIndex = player.SeatIndex
	jd.PlayerInfo = datamsg.MsgPlayerInfo{}
	jd.PlayerInfo.Uid = player.Uid
	jd.PlayerInfo.Name = player.Name
	jd.PlayerInfo.Gold = player.Gold
	jd.PlayerInfo.WinCount = player.WinCount
	jd.PlayerInfo.LoseCount = player.LoseCount
	
	game.sendMsgToAll("SC_PlayerGoIn",jd)
	
}

func (game *Game5GLogic)sendMsgToAll(msgType string,jd interface{}){
	//
	
	data := &datamsg.MsgBase{}
	data.ModeType = "Client"
	data.MsgType = msgType
	
	for _,v := range game.Player{
		if v != nil{
			data.Uid = v.Uid
			data.ConnectId = v.ConnectId
			game.GameAgent.WriteMsgBytes(datamsg.NewMsg1Bytes(data,jd))
		}
	}
	allObserve := game.Observer.Items()
	for _,v := range allObserve{
		if v != nil{
			
			data.Uid = v.(Game5GPlayer).Uid
			data.ConnectId = v.(Game5GPlayer).ConnectId
			game.GameAgent.WriteMsgBytes(datamsg.NewMsg1Bytes(data,jd))
		}
	}
	
}


func (game *Game5GLogic)sendGameInfoToPlayer(player *Game5GPlayer){
	

	//
	data := &datamsg.MsgBase{}
	data.ModeType = "Client"
	data.Uid = player.Uid
	data.ConnectId = player.ConnectId
	data.MsgType = "SC_GameInfo"
	jd := &datamsg.SC_GameInfo{}
	jd.GameInfo = datamsg.MsgGame5GInfo{}
	jd.GameInfo.GameId = game.GameId
	jd.GameInfo.State = game.State
	jd.GameInfo.Time = game.Time
	jd.GameInfo.EveryTime = game.EveryTime
	jd.GameInfo.GameSeatIndex = game.GameSeatIndex
	jd.GameInfo.QiPan = game.QiPan
	
	
	jd.PlayerInfo = make([]datamsg.MsgGame5GPlayerInfo,2)
	jd.ObserveInfo = make([]datamsg.MsgGame5GPlayerInfo,10)
	for _,v := range game.Player{
		if v != nil{
			p1 := datamsg.MsgGame5GPlayerInfo{}
			p1.Uid = v.Uid
			p1.Gold = v.Gold
			p1.LoseCount = v.LoseCount
			p1.Name = v.Name
			p1.PlayerType = v.PlayerType
			p1.SeatIndex = v.SeatIndex
			
			p1.WinCount = v.WinCount
			
			if game.State == Game5GState_Gaming && game.GameSeatIndex == v.SeatIndex{
				//计算用时
				t1 := int((time.Duration(utils.Milliseconde()) - v.OperateStartTime)/1000)-game.EveryTime
				if t1 > 0{
					p1.EveryTime = 0
					p1.Time = v.Time-t1
				}else{
					p1.EveryTime = 0-t1
					p1.Time = v.Time
				}
				
			}else{
				p1.EveryTime = game.EveryTime
				p1.Time = v.Time
			}
			
			
			jd.PlayerInfo = append(jd.PlayerInfo,p1)
		}
	}
	for _,v1 := range game.Observer.Items(){
		if v1 != nil{
			v := v1.(*Game5GPlayer)
			p1 := datamsg.MsgGame5GPlayerInfo{}
			p1.Uid = v.Uid
			p1.Gold = v.Gold
			p1.LoseCount = v.LoseCount
			p1.Name = v.Name
			p1.PlayerType = v.PlayerType
			p1.SeatIndex = v.SeatIndex
			
			p1.WinCount = v.WinCount
			
			jd.ObserveInfo = append(jd.ObserveInfo,p1)
		}
	}
	
	game.GameAgent.WriteMsgBytes(datamsg.NewMsg1Bytes(data,jd))
	
}

//游戏开始
func (game *Game5GLogic)gameStart(){
	//游戏开始
	game.GameSeatIndex = -1
	
	game.Player[0].Time = game.Time
	//Player[0].OperateStartTime = time.Now()
	
	game.Player[1].Time = game.Time
	
	game.State = Game5GState_Gaming
	
	timer.AddCallback(time.Millisecond*3000,game.ChangeGameTurn)
}

//时间到
func (game *Game5GLogic)TimeUp(seatIndex interface{}){
	game.Lock.Lock()
	defer game.Lock.Unlock()
//	Game5GState_Gaming	= 	2	//游戏中
//	Game5GState_Result	= 	3	//结算中
	player := game.Player[seatIndex.(int)]
	if player == nil || player.OperateState == 2{
		return
	}
	si := 0
	if seatIndex == 0{
		si = 1
	}
	game.gameWin(si)
}

//游戏胜利
func (game *Game5GLogic)gameWin(seatIndex int){
	if game.State != Game5GState_Gaming {
		return
	}
	game.State = Game5GState_Result
	
	winplayer := game.Player[seatIndex]
	loseindex := 0
	if seatIndex == 0{
		loseindex = 1
	}
	loseplayer := game.Player[loseindex]
	
	//
	err := db.DbOne.UpdatePlayerWinLose(winplayer.Uid,loseplayer.Uid)
	if err != nil{
		log.Info(err.Error())
		return
		
	}
	//通知玩家数据变化
	
	
	//
	jd := &datamsg.SC_GameOver{}
	jd.WinPlayerSeatIndex = seatIndex
	game.sendMsgToAll("SC_GameOver",jd)
	
	
	//解散房间
	for _,v := range game.Player{
		if v != nil{
			game.GameAgent.Players.Delete(v.Uid)
		}
	}
	allObserve := game.Observer.Items()
	for _,v := range allObserve{
		if v != nil{
			game.GameAgent.Players.Delete(v.(*Game5GPlayer).Uid)
		}
	}
	game.GameAgent.Games.Delete(game.GameId)
	
}




//时间到
func (game *Game5GLogic)createTimeUp(seatIndex int){
	if game.gameTimer != nil{
		game.gameTimer.Cancel()
	}
	game.gameTimer = timer.AddCallback(time.Second*time.Duration(game.Player[game.GameSeatIndex].Time+game.EveryTime),game.TimeUp,seatIndex)
}



func (game *Game5GLogic)ChangeGameTurn(){
	game.Lock.Lock()
	defer game.Lock.Unlock()
	
	game.GameSeatIndex++
	if game.GameSeatIndex >= 2{
		game.GameSeatIndex = 0
	}
	game.Player[game.GameSeatIndex].OperateStartTime = time.Duration(utils.Milliseconde())
	game.Player[game.GameSeatIndex].OperateState = 1
	//计算剩余总时间
	game.createTimeUp(game.GameSeatIndex)
	
	//
	//给所有人发送切换下棋的人信息
	
	jd := &datamsg.SC_ChangeGameTurn{}
	jd.GameSeatIndex = game.GameSeatIndex
	jd.Time = game.Player[game.GameSeatIndex].Time
	jd.EveryTime = game.EveryTime
	game.sendMsgToAll("SC_ChangeGameTurn",jd)
	
}



func (game *Game5GLogic)checkStart(){
	if game.State == Game5GState_Wait {
		for _,v := range game.Player{
			if v == nil{
				return
			}
		}
		
		game.gameStart()
		
		
		//给所有人发送游戏开始信息
		jd := &datamsg.SC_GameStart{}
		jd.GameSeatIndex = game.GameSeatIndex
		game.sendMsgToAll("SC_GameStart",jd)
	}
	
}

//玩家走棋
func (game *Game5GLogic)DoGame5G(playerIndex int,data *datamsg.CS_DoGame5G) (error){
	game.Lock.Lock()
	defer game.Lock.Unlock()
	if playerIndex < 0 || playerIndex >= len(game.Player){
		return errors.New("error playerIndex")
	}
	player := game.Player[playerIndex]
	if data.X < 0 || data.X >= 15 || data.Y < 0 || data.Y >= 15 {
		return errors.New("error x,y")
	}
	
	if game.State != Game5GState_Gaming {
		return errors.New("game is over or no start")
	}
	
	if player.SeatIndex != game.GameSeatIndex || player.OperateState != 1{
		
		return errors.New("no turn you")
	}
	
	if game.QiPan[data.X][data.Y] != -1 {
		return errors.New("here has qizhi")
	}
	
	//走棋成功
	game.QiPan[data.X][data.Y] = player.SeatIndex
	player.OperateState = 2
	
	//计算用时
	t1 := int((time.Duration(utils.Milliseconde()) - player.OperateStartTime)/1000)-game.EveryTime
	if t1 > 0{
		player.Time = player.Time-t1
	}
	
	//给所有人发送走棋
	jd := &datamsg.SC_DoGame5G{}
	jd.GameSeatIndex = player.SeatIndex
	jd.X = data.X
	jd.Y = data.Y
	jd.Time = player.Time
	game.sendMsgToAll("SC_DoGame5G",jd)
	
	//检查是否胜利
	winFlag := game.judgment(data.X,data.Y)
	if winFlag != -1 {
		game.gameWin(winFlag)
	}else{
		game.Lock.Unlock()
		game.ChangeGameTurn()
		game.Lock.Lock()
	}
	
	
	//SC_DoGame5G
	
	
	//SC_DoGame5G
	return nil
}

//玩家进入
func (game *Game5GLogic)GoIn(player *Game5GPlayer) (*Game5GPlayer,error){
	game.Lock.Lock()
	defer game.Lock.Unlock()
	
	//游戏结束
	if game.State >= Game5GState_Result{
		
		return nil,errors.New("game over!")
	}
	
	//玩家进入
	for k,v := range game.WillPlayGamePlayerUid{
		if v == player.Uid{
			player.SeatIndex = k
			player.PlayerType = 1
			
			if game.Player[k] != nil{
				game.Player[k].ConnectId = player.ConnectId
				//给其他玩家发送这个玩家断线重连
				
			}else{
				player.Game = game
				game.notifyAllPlayerGoIn(player)
				game.Player[k] = player
			}
			
			
			game.sendGameInfoToPlayer(game.Player[k])
			
			
			//检查玩家是否到齐  游戏能否开始
			game.checkStart()
			
			return game.Player[k],nil
			
		}
	}
	
	//旁观者进入
	player.PlayerType = 2
	player.SeatIndex = -2
	game.notifyAllPlayerGoIn(player)
	game.Observer.Set(player.Uid,player)
	game.sendGameInfoToPlayer(player)
	return player,nil
	
}


//玩家退出
func (game *Game5GLogic)GoOut(player *Game5GPlayer)bool{
	game.Lock.Lock()
	defer game.Lock.Unlock()
	
	//游戏结束
	if game.State >= Game5GState_Result{
		return true
	}
	
	//玩家
	if player.PlayerType == 1{
		
		game.Player[player.SeatIndex] = nil
		//给所有人发送玩家离开
		jd := &datamsg.SC_PlayerGoOut{}
		jd.Uid = player.Uid
		game.sendMsgToAll("SC_PlayerGoOut",jd)
		
		
		wi := 0
		if player.SeatIndex == 0{
			wi = 1
		}
		game.gameWin(wi)
		return true
	}
	
	//观察者
	game.Observer.Delete(player.Uid)
	
	//给所有人发送玩家离开
	jd := &datamsg.SC_PlayerGoOut{}
	jd.Uid = player.Uid
	game.sendMsgToAll("SC_PlayerGoOut",jd)
	
	return true
}



//玩家掉线
func (game *Game5GLogic)Disconnect(player *Game5GPlayer)bool{
	game.Lock.Lock()
	defer game.Lock.Unlock()
	
	//游戏结束
	if game.State >= Game5GState_Result{
		return  true
	}
	
	//玩家
	if player.PlayerType == 1{
		
		//玩家掉线中标志
		
		return true
	}
	
	//观察者
	game.Observer.Delete(player.Uid)
	
	//给所有人发送玩家离开
	jd := &datamsg.SC_PlayerGoOut{}
	jd.Uid = player.Uid
	game.sendMsgToAll("SC_PlayerGoOut",jd)
	return true
	
}











func (game *Game5GLogic)judgment(x int, y int)int{
    winFlag := -1
	data := game.QiPan
	for i := 0; i != 5; i++{
	    if 	(y-i>=0 && y-i+4<15 &&   
            data[x][y-i]==data[x][y-i+1] &&    // 横   
            data[x][y-i]==data[x][y-i+2] &&   
            data[x][y-i]==data[x][y-i+3] &&  
            data[x][y-i]==data[x][y-i+4]) ||   
           	(x-i>=0 && x-i+4<15 &&             // 竖   
            data[x-i][y]==data[x-i+1][y] &&   
            data[x-i][y]==data[x-i+2][y] &&   
            data[x-i][y]==data[x-i+3][y] &&   
            data[x-i][y]==data[x-i+4][y]) ||   
           	(x-i>=0 && y-i>=0 && x-i+4<15 && y-i+4<15 &&  // 左向右斜  
            data[x-i][y-i]==data[x-i+1][y-i+1] &&  
            data[x-i][y-i]==data[x-i+2][y-i+2] &&   
            data[x-i][y-i]==data[x-i+3][y-i+3] &&   
            data[x-i][y-i]==data[x-i+4][y-i+4]) ||   
           	(x-i>=0 && y+i<15 && x-i+4<15 && y+i-4>=0 &&  // 右向左斜  
            data[x-i][y+i]==data[x-i+1][y+i-1] &&   
            data[x-i][y+i]==data[x-i+2][y+i-2] &&   
            data[x-i][y+i]==data[x-i+3][y+i-3] &&   
            data[x-i][y+i]==data[x-i+4][y+i-4]){
		       winFlag = data[x][y]
		       break
	   		}
	}
    return winFlag
}













var g_GameId = 10000
var g_GameId_lock = new(sync.Mutex)

//创建一个新的游戏ID
func GetNewGameId() int{
	g_GameId_lock.Lock()
	defer g_GameId_lock.Unlock()
	
	g_GameId++
	return g_GameId
}

//创建游戏
func NewGame5GLogic(ga *Game5GAgent,p1Id int,p2Id int,time int,everytime int) *Game5GLogic{
	ng := &Game5GLogic{}
	ng.GameId = GetNewGameId()
	ng.WillPlayGamePlayerUid[0] = p1Id
	ng.WillPlayGamePlayerUid[1] = p2Id
	ng.GameAgent = ga
	ng.Time = time
	ng.EveryTime = everytime
	
	ng.Init()
	return ng
}

//游戏逻辑管理器