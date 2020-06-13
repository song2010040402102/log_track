package stat

import (
	"common"
	"data"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
	"util"
)

func AllOnlineTimeHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	return onlineTimeHandler(gameId, startDate, endDate, channel, device, false)
}

func NewOnlineTimeHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	return onlineTimeHandler(gameId, startDate, endDate, channel, device, true)
}

func onlineTimeHandler(gameId int32, startDate, endDate string, channel string, device string, bNew bool) *ItemResult {
	userTimes := getOnlineTime(gameId, startDate, endDate, channel, device, bNew)
	var title string
	if bNew {
		title = fmt.Sprintf("新增用户在线时长统计(%s ~ %s)", startDate, endDate)
	} else {
		title = fmt.Sprintf("所有用户在线时长统计(%s ~ %s)", startDate, endDate)
	}
	heads := []string{"日期", "渠道", "设备", "<1m用户数", "占比", "1~3m用户数", "占比", "3~6m用户数", "占比", "6~10m用户数", "占比",
		"10~30m用户数", "占比", "30m~1h用户数", "占比", "1~2h用户数", "占比", "2~4h用户数", "占比", "4~6h用户数", "占比",
		"6~8h用户数", "占比", "8~10h用户数", "占比", ">=10h用户数", "占比", "平均时长"}
	items := make([][]string, 0, len(userTimes))
	for i, mt := range userTimes {
		rows := make([]string, 0, len(heads))
		if !bNew && i == len(userTimes)-1 {
			rows = append(rows, "平均")
		} else {
			rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
		}
		rows = append(rows, channel)
		rows = append(rows, device)
		tc := [12]int32{}
		sum := int32(0)
		for _, t := range mt {
			if !bNew && i == len(userTimes)-1 {
				t = t / int32(len(userTimes)-1)
			}
			if t < 60 {
				tc[0]++
			} else if t < 180 {
				tc[1]++
			} else if t < 360 {
				tc[2]++
			} else if t < 600 {
				tc[3]++
			} else if t < 1800 {
				tc[4]++
			} else if t < 3600 {
				tc[5]++
			} else if t < 7200 {
				tc[6]++
			} else if t < 14400 {
				tc[7]++
			} else if t < 21600 {
				tc[8]++
			} else if t < 28800 {
				tc[9]++
			} else if t < 36000 {
				tc[10]++
			} else {
				tc[11]++
			}
			sum += t
		}
		for _, v := range tc {
			rows = append(rows, strconv.Itoa(int(v)))
			if len(mt) > 0 {
				rows = append(rows, fmt.Sprintf("%.2f%%", float32(v)*100/float32(len(mt))))
			} else {
				rows = append(rows, "0.00%")
			}
		}
		if len(mt) > 0 {
			rows = append(rows, strconv.Itoa(int(sum)/len(mt)))
		} else {
			rows = append(rows, "0")
		}
		items = append(items, rows)
	}
	return &ItemResult{
		TTable: NewTable(title, heads, items),
	}
}

func getOnlineTime(gameId int32, startDate, endDate string, channel string, device string, bNew bool) []map[string]int32 {
	var sumTimes map[string]int32
	if !bNew {
		sumTimes = make(map[string]int32)
	}
	var mapTimes map[string]int32
	var userTimes []map[string]int32
	data.GetLogData(gameId, "Logout", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos == 0 {
			mapTimes = make(map[string]int32)
		} else if pos > 0 {
			if len(s) < 11 {
				logs.Error("getOnlineTime data error!")
				return false
			}
			if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
				return true
			}
			t, _ := strconv.ParseInt(s[10], 10, 32)
			if bNew {
				registerTime, _ := strconv.ParseInt(s[6], 10, 32)
				if util.Ts2date(util.Date2ts(startDate)+int64(day*86400)) == util.Ts2date(registerTime) {
					mapTimes[s[0]] += int32(t)
				}
			} else {
				mapTimes[s[0]] += int32(t)
				sumTimes[s[0]] += int32(t)
			}
		} else {
			userTimes = append(userTimes, mapTimes)
		}
		return true
	})
	if !bNew {
		userTimes = append(userTimes, sumTimes)
	}
	return userTimes
}
