package utils

import (
	"strconv"
	"strings"
	"time"
)

func Milliseconde() int64 {
	return time.Now().UnixNano() / 1000000
}

//func GetSmallTypeByItemType(itemType int) int {
//	if itemType < 1100 {
//		return 1
//	} else if itemType < 1200 {
//		return 2
//	} else if itemType < 1300 {
//		return 3
//	} else if itemType < 1400 {
//		return 4
//	} else if itemType < 1500 {
//		return 5
//	}
//	return 1
//}

//检查时间是否过期
func CheckTimeIsExpiry(timestr string) bool {
	nowtime, _ := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"))
	time, _ := time.Parse("2006-01-02 15:04:05", timestr)
	if time.Before(nowtime) {
		return true
	} else {
		return false
	}
}

//毫秒 从t1时间开始计算睡眠时间   睡眠sleeptime毫秒
func MySleep(t1 int64, sleeptime int64) {
	t2 := Milliseconde()
	t3 := t2 - t1
	if t3 >= sleeptime {
		return
	} else {
		d := (time.Duration)(sleeptime-t3) * time.Millisecond
		//fmt.Println("-----d:",int64(d))
		time.Sleep(d)
	}
}

func SplitStringToIntArray(str string) []int {

	re := make([]int, 0)

	strs := strings.Split(str, ",")
	if len(strs) > 0 {

		for i := 0; i < len(strs); i++ {
			mailid, err := strconv.Atoi(strs[i])
			if err != nil {
				//index++
				continue
			}
			re = append(re, mailid)
		}

	}

	return re

}

func GetMinMaxUidStr(uid1 int, uid2 int) string {
	minuid := uid1
	maxuid := uid2
	if minuid > maxuid {
		minuid = uid2
		maxuid = uid1
	}
	twouidstr := strconv.Itoa(minuid) + "_" + strconv.Itoa(maxuid)

	return twouidstr
}
func GetMin(uid1 int, uid2 int) int {
	if uid1 > uid2 {
		return uid2
	}

	return uid1
}
func GetMax(uid1 int, uid2 int) int {
	if uid1 > uid2 {
		return uid1
	}

	return uid2
}

func SplitStringToIntMap(str string) map[int]interface{} {

	re := make(map[int]interface{})

	strs := strings.Split(str, ",")
	if len(strs) > 0 {

		for i := 0; i < len(strs); i++ {
			mailid, err := strconv.Atoi(strs[i])
			if err != nil {
				//index++
				continue
			}
			re[mailid] = true
		}

	}

	return re

}
