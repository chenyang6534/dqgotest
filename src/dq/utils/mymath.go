package utils

import (
	"strconv"
	"strings"
	"time"
)

func Milliseconde() int64 {
	return time.Now().UnixNano() / 1000000
}

//毫秒
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
