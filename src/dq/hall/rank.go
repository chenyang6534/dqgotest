package hall

import (
	"dq/conf"
	"dq/datamsg"
	"dq/db"
	"dq/log"
	"dq/utils"
	"strconv"
	//"strings"
	"dq/timer"
	//"math"
	"sort"
	"sync"
	"time"
)

var (
	rank      = &Rank{Lock: new(sync.Mutex)}
	RankCount = 1000
)

//type RankNode struct {
//	Uid    int
//	Score  int
//	Name   string
//	Avatar string
//}

type Rank struct {
	Lock *sync.Mutex

	RankList     []datamsg.RankNodeInfo //排行榜缓存
	RankNumOfUid *utils.BeeMap          //玩家排名缓存

	RankMap *utils.BeeMap //需要排行的数据

	RankTimer     *timer.Timer
	SeasonIdIndex int

	SeasonOverTimerCheck *timer.Timer
	DoSeason             sync.WaitGroup
}

func GetRank() *Rank {
	return rank
}

type RankSlice []datamsg.RankNodeInfo

func (s RankSlice) Len() int           { return len(s) }
func (s RankSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s RankSlice) Less(i, j int) bool { return s[i].Score > s[j].Score }

func GetCurSeasonIdIndex() int {
	seasons := conf.GetSeasonConfig().Seasons
	for _, v := range seasons {
		//
		if utils.CheckTimeIsExpiry(v.StartTime) == true && utils.CheckTimeIsExpiry(v.EndTime) == false {

			return v.IdIndex
		}

	}

	return -1

}

func GetSeasonRewardFromRank(seasonidindex int, ranknum int) int {
	seasons := conf.GetSeasonConfig().Seasons
	for _, v := range seasons {
		//
		if v.IdIndex == seasonidindex {
			for _, v1 := range v.RewardList {
				if ranknum >= v1.RankStart && ranknum <= v1.RankEnd {

					return v1.Gold
				}
			}

			return -1

		}

	}
	return -1
}

//派发上赛季奖励
func (rank *Rank) GaveReward(seasonidindex int) {

	ranklist := make([]datamsg.RankNodeInfo, 0)
	db.DbOne.GetRankInfo(&ranklist, seasonidindex)

	log.Info("---seasonidindex--%d--size-%d", seasonidindex, len(ranklist))

	//排序
	sort.Sort(RankSlice(ranklist))

	ismail := false

	for k, v := range ranklist {
		//log.Info("---rank--%d", v.Rewardgold)
		ranknum := k + 1
		if v.Rewardgold == 0 { //没有发放奖励
			//发奖励
			rewardgold := GetSeasonRewardFromRank(seasonidindex, ranknum)
			//log.Info("---rewardgold--%d---rank:%d", rewardgold, ranknum)

			if rewardgold > 0 {
				//发送奖励邮件
				mailp := &datamsg.MailInfo{}
				mailp.SendName = "宝石五子棋"
				mailp.Title = "赛季奖励"
				mailp.Content = "恭喜你在本赛季获得第" + strconv.Itoa(ranknum) + "名,特此奖励!"
				mailp.RecUid = v.Uid
				mailp.Date = time.Now().Format("2006-01-02")
				tt := conf.RewardsConfig{Type: 1, Count: rewardgold, Time: -1}
				mailp.Reward = append(mailp.Reward, tt)
				mailp.ReadState = 0
				mailp.GetState = 0
				db.DbOne.WritePrivateMailInfo(v.Uid, mailp)
			}

			//
			ranklist[k].Rewardgold = rewardgold

			ismail = true

		}
	}
	//入库
	if ismail {
		db.DbOne.WriteRankInfo(ranklist, seasonidindex)

		//赛季结束后的积分变换
		db.DbOne.ResetAllPlayerSeasonScore()
	}

}

func (rank *Rank) Init() {
	rank.RankMap = utils.NewBeeMap()
	rank.RankNumOfUid = utils.NewBeeMap()
	rank.RankList = make([]datamsg.RankNodeInfo, 0)

	rank.SeasonIdIndex = GetCurSeasonIdIndex()

	//检查上个赛季的奖励是否发放完毕
	rank.GaveReward(rank.SeasonIdIndex - 1)

	//GetRankInfo

	//从数据库读取数据
	db.DbOne.GetRankInfo(&rank.RankList, rank.SeasonIdIndex)

	//排序
	sort.Sort(RankSlice(rank.RankList))

	for k, v := range rank.RankList {
		//log.Info("----11-name:%s---score:%d---rank%d", v.Name, v.Score, k)
		rank.RankNumOfUid.Set(v.Uid, k+1)
	}
	//map
	rank.ListToMap()

	//--
	rank.RankTimer = timer.AddRepeatCallback(time.Second*30, rank.Sort)

	rank.SeasonOverTimerCheck = timer.AddRepeatCallback(time.Second*30, rank.CheckSeasonOver)
	//t.repeat = true

}

//设置变化后的值
func (rank *Rank) SetValue(data datamsg.RankNodeInfo) {
	rank.Lock.Lock()
	defer rank.Lock.Unlock()

	rank.RankMap.Set(data.Uid, data)

}

//当赛季结束时的操作
func (rank *Rank) DoSeasonOver() {

	curseason := GetCurSeasonIdIndex()

	rank.Sort()

	db.DbOne.WriteRankInfo(rank.RankList, rank.SeasonIdIndex)
	rank.SeasonIdIndex = curseason

	//检查上个赛季的奖励是否发放完毕
	rank.GaveReward(rank.SeasonIdIndex - 1)

	//赛季结束后的积分变换
	//db.DbOne.ResetAllPlayerSeasonScore()

	//删除所有缓存数据
	rank.RankList = make([]datamsg.RankNodeInfo, 0)
	rank.RankNumOfUid.DeleteAll()
	rank.RankMap.DeleteAll()

	rank.DoSeason.Done()

	log.Info("---DoSeasonOver---")
}

//检查是否赛季到期
func (rank *Rank) CheckSeasonOver() {
	log.Info("---CheckSeasonOver---")
	curseason := GetCurSeasonIdIndex()
	//赛季到期
	if curseason > rank.SeasonIdIndex {

		rank.DoSeason.Add(1)
		go rank.DoSeasonOver()

	}
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
		rank.RankNumOfUid.Set(v.Uid, k+1)

		rank.RankMap.Set(v.Uid, v)

	}

	t2 := utils.Milliseconde()
	log.Info("----ranksort--time:%d", (t2 - t1))

}

//获取前3个赛季的最高排名
func (rank *Rank) GetMaxRankNum3(uid int) int {

	maxrank := 10000
	for i := 1; i <= 3; i++ {
		index := rank.SeasonIdIndex - i
		ranknum := db.DbOne.GetRankNum(index, uid)
		if ranknum < maxrank {
			maxrank = ranknum
		}

	}

	return maxrank

}

//
func (rank *Rank) RankInfo(start int, end int, uid int) *datamsg.SC_RankInfo {

	rankinfo := &datamsg.SC_RankInfo{}

	seasons := conf.GetSeasonConfig().Seasons
	for _, v := range seasons {
		//
		if v.IdIndex == rank.SeasonIdIndex {
			rankinfo.SeasonInfo = v
			break
		}
	}

	rankinfo.MyRank = RankCount + 1
	if start >= end || start < 0 {
		return rankinfo
	}

	rank.Lock.Lock()
	defer rank.Lock.Unlock()

	ranklen := len(rank.RankList)
	if ranklen <= 0 {
		return rankinfo
	}

	for i := start; i < end; i++ {

		if i < ranklen {
			msg := datamsg.RankNodeMessage{rank.RankList[i], i + 1}
			rankinfo.Ranks = append(rankinfo.Ranks, msg)
		}
	}

	//自己的排名
	if rank.RankNumOfUid.Check(uid) == true {
		rankinfo.MyRank = (rank.RankNumOfUid.Get(uid)).(int)
	}

	//	if rank.RankList[ranklen-1].Score <= myscore {
	//		rankinfo.MyRank = rank.TwoPointFindRankNum(myscore, uid)
	//		log.Info("----------myrank:%d---myscore:%d--uid:%d", rankinfo.MyRank, myscore, uid)
	//	}

	return rankinfo

}

func (rank *Rank) WriteDB() {

	if rank.RankTimer != nil {
		rank.RankTimer.Cancel()
	}

	rank.Sort()

	rank.Lock.Lock()
	defer rank.Lock.Unlock()

	db.DbOne.WriteRankInfo(rank.RankList, rank.SeasonIdIndex)

	rank.DoSeason.Wait()

}

func (rank *Rank) ListToMap() {
	if rank.RankMap != nil {
		rank.RankMap.DeleteAll()
	}
	//读取数组中的数据
	for _, v := range rank.RankList {
		rank.RankMap.Set(v.Uid, v)
	}

	log.Info("---rankmap size:%d", rank.RankMap.Size())
}
