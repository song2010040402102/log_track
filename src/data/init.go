package data

import (
	"fmt"
	"time"
)

const DAILY_DATA_TIME string = "01:00:00" //每天自动获取数据时间

func Init() {
	InitUserData()
	InitTributeData()
	AutoGetData()
}

func Flush() {
	FlushUserData()
	FlushTributeData()
}

func AutoGetData() {
	timer := time.NewTimer(getRemainSecond(DAILY_DATA_TIME))
	go func(t *time.Timer) {
		for {
			<-t.C
			AutoGetLogData()
			AutoGetTributeData()
			t.Reset(getRemainSecond(DAILY_DATA_TIME))
		}
	}(timer)
}

func getRemainSecond(times string) time.Duration {
	t := time.Now()
	n1 := t.Hour()*3600 + t.Minute()*60 + t.Second()

	h, m, s := 0, 0, 0
	fmt.Sscanf(times, "%d:%d:%d", &h, &m, &s)
	n2 := h*3600 + m*60 + s

	if n2 > n1 {
		return time.Duration(n2-n1) * time.Second
	} else {
		return time.Duration(n2-n1+86400) * time.Second
	}
}
