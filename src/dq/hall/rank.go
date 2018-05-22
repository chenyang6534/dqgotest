package hall

import (
	//"dq/conf"
	"dq/datamsg"
	"dq/db"
	"dq/log"
	"dq/utils"
	//"strconv"
	//"strings"
	"dq/timer"
	"sort"
	"sync"
	"time"
)

var (
	rank      = &Rank{Lock: new(sync.Mutex)}
	RankCount = 500
)

//type RankNode struct {
//	Uid    int
//	Score  int
//	Name   string
//	Avatar string
//}

type Rank struct {
	Lock *sync.Mutex

	RankList  []datamsg.RankNodeInfo
	RankMap   *utils.BeeMap
	RankTimer *timer.Timer
}

func GetRank() *Rank {
	return rank
}

type RankSlice []datamsg.RankNodeInfo

func (s RankSlice) Len() int           { return len(s) }
func (s RankSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s RankSlice) Less(i, j int) bool { return s[i].Score > s[j].Score }

func (rank *Rank) Init() {
	rank.RankMap = utils.NewBeeMap()
	rank.RankList = make([]datamsg.RankNodeInfo, 0)

	//从数据库读取数据
	db.DbOne.GetRankInfo(&rank.RankList)

	//排序
	sort.Sort(RankSlice(rank.RankList))

	for k, v := range rank.RankList {
		log.Info("----11-name:%s---score:%d---rank%d", v.Name, v.Score, k)
	}
	//map
	rank.ListToMap()

	//--
	rank.RankTimer = timer.AddRepeatCallback(time.Second*30, rank.Sort)
	//t.repeat = true

}

//设置变化后的值
func (rank *Rank) SetValue(data datamsg.RankNodeInfo) {
	rank.Lock.Lock()
	defer rank.Lock.Unlock()

	rank.RankMap.Set(data.Uid, data)

}

//重新排序
func (rank *Rank) Sort() {
	rank.Lock.Lock()
	defer rank.Lock.Unlock()

	t1 := utils.Milliseconde()

	size := rank.RankMap.Size()
	log.Info("---rank size:%d", size)
	if size > RankCount {
		size = RankCount
	}

	templist := make([]datamsg.RankNodeInfo, 0)
	//读取map中的数据
	items := rank.RankMap.Items()
	for _, v := range items {
		templist = append(templist, v.(datamsg.RankNodeInfo))
	}
	//清除map
	if rank.RankMap != nil {
		rank.RankMap.DeleteAll()
	}

	//	for k, v := range templist {
	//		log.Info("----11-name:%s---score:%d---rank%d", v.Name, v.Score, k)
	//	}

	//排序
	sort.Sort(RankSlice(templist))
	rank.RankList = make([]datamsg.RankNodeInfo, size)

	//rank.RankList = append(templist[:size])

	//count := 0
	//RankCount
	for k, v := range templist {
		//log.Info("----ranksort-name:%s---score:%d---rank%d", v.Name, v.Score, k)
		//count++
		if k >= RankCount {
			break
		}
		rank.RankList[k] = v
		rank.RankMap.Set(v.Uid, v)

	}

	t2 := utils.Milliseconde()
	log.Info("----ranksort--time:%d", (t2 - t1))

}

//
func (rank *Rank) RankInfo(start int, end int) *datamsg.SC_RankInfo {
	rankinfo := &datamsg.SC_RankInfo{}
	if start >= end || start < 0 {
		return rankinfo
	}

	rank.Lock.Lock()
	defer rank.Lock.Unlock()

	ranklen := len(rank.RankList)

	for i := start; i < end; i++ {

		if i < ranklen {
			msg := datamsg.RankNodeMessage{rank.RankList[i], i + 1}
			rankinfo.Ranks = append(rankinfo.Ranks, msg)
		}

	}

	return rankinfo

}

func (rank *Rank) WriteDB() {

	if rank.RankTimer != nil {
		rank.RankTimer.Cancel()
	}

	rank.Sort()

	rank.Lock.Lock()
	defer rank.Lock.Unlock()

	db.DbOne.WriteRankInfo(rank.RankList)

}

func (rank *Rank) ListToMap() {
	if rank.RankMap != nil {
		rank.RankMap.DeleteAll()
	}
	//读取数组中的数据
	for _, v := range rank.RankList {
		rank.RankMap.Set(v.Uid, v)
	}
}
