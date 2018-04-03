package hall
import(
	"dq/network"
	
	"dq/log"
	"dq/datamsg"
	"net"
	"encoding/json"
	//"strconv"
	//"time"
	"dq/db"
	"dq/utils"
)
type HallAgent struct {
	conn     network.Conn
	
	
	userdata string
	
	handles  map[string]func(data *datamsg.MsgBase)
	
	//
	closeFlag *utils.BeeVar
	
	//五指棋 玩家匹配表[id] = 
	serchPoolFor5G	*utils.BeeMap
}


type serchInfo struct{
	ConnectId int
	Uid			int
	
	Time		int	//游戏总时间 s
	EveryTime	int //游戏每一步棋时间 s
}


func (a *HallAgent) GetConnectId() int{
	
		return 0
}
func (a *HallAgent) GetModeType() string{
	return ""
}

func (a *HallAgent) Init() {
	
	a.serchPoolFor5G = utils.NewBeeMap()
	a.closeFlag = utils.NewBeeVar(false)
	
	
	a.handles = make(map[string]func(data *datamsg.MsgBase))
	
	a.handles["GetInfo"] = a.DoGetInfoData
	
	a.handles["CS_QuickGame"] = a.DoQuickGameData
	a.handles["CS_QuickGameExit"] = a.DoQuickGameExitData
	
	//玩家断线
	a.handles["Disconnect"] = a.DoDisConnectData
	
}

func (a *HallAgent) DoQuickGameData(data *datamsg.MsgBase) {
//	log.Info("--modeType:"+data.ModeType)
//	log.Info("--ConnectId:"+strconv.Itoa(data.ConnectId))
//	log.Info("--MsgId:"+strconv.Itoa(data.MsgId))
	
	sinfo := &serchInfo{}
	sinfo.ConnectId = data.ConnectId
	sinfo.Uid = data.Uid
	sinfo.Time = 60*20	//20分钟
	sinfo.EveryTime = 30 //30秒
	
	a.serchPoolFor5G.Set(data.Uid,sinfo)
	
	
	//回复客户端 搜寻中
	playerinfo := &datamsg.MsgPlayerInfo{}
	err := db.DbOne.GetPlayerInfo(data.Uid,playerinfo)
	if err == nil{
		data.ModeType = "Client"
		data.MsgType = "SC_SerchPlayer"
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data,nil))
	}
	
}

func (a *HallAgent) DoDisConnectData(data *datamsg.MsgBase) {
	
	log.Info("----DoDisConnectData uid:%d--",data.Uid)
	a.serchPoolFor5G.Delete(data.Uid)
	
}

func (a *HallAgent) DoQuickGameExitData(data *datamsg.MsgBase) {

	if a.serchPoolFor5G.Check(data.Uid){
		a.serchPoolFor5G.Delete(data.Uid)
		
		//回复客户端 搜寻中
		data.ModeType = "Client"
		data.MsgType = "SC_QuickGameExit"
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data,nil))
		
		return
	}
	
}


func (a *HallAgent) DoGetInfoData(data *datamsg.MsgBase) {

	
	//回复客户端
	playerinfo := &datamsg.MsgPlayerInfo{}
	err := db.DbOne.GetPlayerInfo(data.Uid,playerinfo)
	if err == nil{
		data.ModeType = "Client"
		data.MsgType = "SC_MsgHallInfo"
		jd := datamsg.SC_MsgHallInfo{}
		jd.PlayerInfo = *playerinfo
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data,jd))
	}
	
	
	
	
}



func (a *HallAgent) Update() {
	
	//500毫秒循环一次
	oneUpdateTime := 500
	
	for {
		t1 := utils.Milliseconde()
		if a.closeFlag.Get() == true{
			return
		}
		serchPlayer := a.serchPoolFor5G.Items()
		
		//算法忽略
		
		//
		size := len(serchPlayer)
		if size <= 1 {
			utils.MySleep(t1,int64(oneUpdateTime))
			continue
		}
		fight := [2]int{}
		i := 0
		for k,_ := range serchPlayer{
			fight[i] = k.(int)
			delete(serchPlayer,k)
			i++;
			if i >= 2{
				break
			}
		}
		
		//算法结束
		if a.closeFlag.Get() == true{
			return
		}
		
		
		p1 := a.serchPoolFor5G.Get(fight[0])
		p2 := a.serchPoolFor5G.Get(fight[1])
		if p1 != nil && p2 != nil{
			a.serchPoolFor5G.Delete(fight[0])
			a.serchPoolFor5G.Delete(fight[1])
			//创建一个游戏
			a.CreateGame(p1.(*serchInfo),p2.(*serchInfo))
		}
		
		utils.MySleep(t1,int64(oneUpdateTime))
		
	}
}


func (a *HallAgent) CreateGame(arg *serchInfo,arg1 *serchInfo){
	
	//通知游戏 开始一局新游戏
	data := &datamsg.MsgBase{}
	data.ModeType = "Game5G"
	data.Uid = 0
	data.MsgType = "NewGame"
	jd := make(map[string]interface{})
	jd["player1"] = arg.Uid //p1
	jd["player2"] = arg1.Uid //p2
	jd["player1ConnectId"] = arg.ConnectId //p1
	jd["player2ConnectId"] = arg1.ConnectId //p2
	jd["time"] = arg.Time
	jd["everytime"] = arg.EveryTime
	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data,jd))
}




func (a *HallAgent) Run() {
	
	a.Init()
	
	go a.Update()
	
	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err)
			break
		}
		
		go a.doMessage(data)
		
	}
}

func (a *HallAgent) doMessage(data []byte) {
	log.Info("----------Hall----readmsg---------")
	h1 := &datamsg.MsgBase{}
	err := json.Unmarshal(data,h1)
	if err != nil{
		log.Info("--error")
	}else{
		
		log.Info("--MsgType:"+h1.MsgType)
		if f,ok := a.handles[h1.MsgType];ok{
			f(h1)
		}
		
		
		
	}

}



func (a *HallAgent) OnClose() {
	

}

func (a *HallAgent) WriteMsg(msg interface{}) {

}
func (a *HallAgent) WriteMsgBytes(msg []byte) {

		err := a.conn.WriteMsg(msg)
		if err != nil {
			log.Error("write message  error: %v", err)
		}
}
func (a *HallAgent) RegisterToGate() {
	t2 := datamsg.MsgRegisterToGate{
		ModeType	: datamsg.HallMode,
	}
	
	t1 := datamsg.MsgBase{
		ModeType	: datamsg.GateMode,
		MsgType		:"Register",
	}
	
	a.WriteMsgBytes(datamsg.NewMsg1Bytes(&t1,&t2))
	
//	temp,err := json.Marshal(t2)
//	t1.JsonData	= string(temp)
//	data,err := json.Marshal(t1)
//	if err != nil{
//		log.Info("-------json.Marshal:"+err.Error())
//		return
//	}
	
//	a.WriteMsgBytes(data)
	
}


func (a *HallAgent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *HallAgent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *HallAgent) Close() {
	a.closeFlag.Set(true)
	a.conn.Close()
}

func (a *HallAgent) Destroy() {
	a.conn.Destroy()
}