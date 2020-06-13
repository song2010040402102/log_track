package stat

import (
	"common"
	"data"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
	"util"
)

func AllPlayHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	return playHandler(gameId, startDate, endDate, channel, device, false)
}

func NewPlayHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	return playHandler(gameId, startDate, endDate, channel, device, true)
}

func playHandler(gameId int32, startDate, endDate string, channel string, device string, bNew bool) *ItemResult {
	plays := getPlay(gameId, startDate, endDate, channel, device, bNew)
	var daus []map[string]bool
	var creates []map[string]bool
	if bNew {
		creates, _, _, _ = GetCreate(gameId, startDate, endDate, channel, device)
	} else {
		daus = GetDAU(gameId, startDate, endDate, channel, device)
	}
	if len(creates) > 0 && len(creates) != len(plays) || len(daus) > 0 && len(daus) != len(plays) {
		logs.Error("playHandler, data length error!")
		return nil
	}
	var title string
	if bNew {
		title = fmt.Sprintf("新增用户打牌统计(%s ~ %s)", startDate, endDate)
	} else {
		title = fmt.Sprintf("所有用户打牌统计(%s ~ %s)", startDate, endDate)
	}
	heads := []string{"日期", "渠道", "设备", "0局", "占比", "1局", "占比", "2~5局", "占比", "6~10局", "占比", ">10局", "占比"}
	items := make([][]string, 0, len(plays))
	for i, mp := range plays {
		rows := make([]string, 0, len(heads))
		rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
		rows = append(rows, channel)
		rows = append(rows, device)
		tc := [5]int32{}
		for _, n := range mp {
			if n == 1 {
				tc[1]++
			} else if n > 1 && n < 6 {
				tc[2]++
			} else if n > 5 && n < 11 {
				tc[3]++
			} else if n > 10 {
				tc[4]++
			}
		}
		sum := int32(0)
		if bNew {
			sum = int32(len(creates[i]))
		} else {
			sum = int32(len(daus[i]))
		}
		tc[0] = sum - tc[1] - tc[2] - tc[3] - tc[4]
		for _, v := range tc {
			rows = append(rows, strconv.Itoa(int(v)))
			if sum > 0 {
				rows = append(rows, fmt.Sprintf("%.2f%%", float32(v)*100/float32(sum)))
			} else {
				rows = append(rows, "0.00%")
			}
		}
		items = append(items, rows)
	}
	return &ItemResult{
		TTable: NewTable(title, heads, items),
	}
}

func getPlay(gameId int32, startDate, endDate string, channel string, device string, bNew bool) []map[string]int32 {
	var mapPlay map[string]int32
	var plays []map[string]int32
	data.GetLogData(gameId, "RoomFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos == 0 {
			mapPlay = make(map[string]int32)
		} else if pos > 0 {
			if len(s) < 11 {
				logs.Error("getPlay data error!")
				return false
			}
			if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
				return true
			}
			if bNew {
				registerTime, _ := strconv.ParseInt(s[6], 10, 32)
				if util.Ts2date(util.Date2ts(startDate)+int64(day*86400)) == util.Ts2date(registerTime) {
					mapPlay[s[0]]++
				}
			} else {
				mapPlay[s[0]]++
			}
		} else {
			plays = append(plays, mapPlay)
		}
		return true
	})
	return plays
}
