package utils

import(
	"time"
)

func Milliseconde() int64{
	return time.Now().UnixNano()/1000000
}


//毫秒
func MySleep(t1 int64,sleeptime int64){
	t2 := Milliseconde()
	t3 := t2-t1
	if t3 >= sleeptime{
		return
	}else{
		d := (time.Duration)(sleeptime-t3)*time.Millisecond
		//fmt.Println("-----d:",int64(d))
		time.Sleep( d)
	}
}