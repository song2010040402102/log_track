package stat

import (
	"common"
	"data"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
	"util"
)

func SysUserHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	_, createChild, noShareTour, noShareAuth := GetCreate(gameId, startDate, endDate, channel, device)
	dau := GetDAU(gameId, startDate, endDate, channel, device)
	sumAuth := getSumAuth(gameId, startDate, endDate, channel, device)
	if len(createChild) != len(noShareTour) || len(noShareTour) != len(noShareAuth) || len(noShareAuth) != len(dau) || len(dau) != len(sumAuth) {
		logs.Error("SysUserHandler, data length error!")
		return nil
	}
	heads := []string{"日期", "渠道", "设备", "不含分享创建数", "试玩数", "授权数", "试玩转授权数", "裂变率"}
	items := make([][]string, 0, len(dau))
	for i, v := range dau {
		rows := make([]string, 0, len(heads))
		rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
		rows = append(rows, channel)
		rows = append(rows, device)
		rows = append(rows, strconv.Itoa(len(noShareTour[i])+len(noShareAuth[i])))
		rows = append(rows, strconv.Itoa(len(noShareTour[i])))
		rows = append(rows, strconv.Itoa(len(noShareAuth[i])))
		sum := 0
		for k, _ := range sumAuth[i] {
			if _, ok := noShareTour[i][k]; ok {
				sum++
			}
		}
		rows = append(rows, strconv.Itoa(sum))
		if len(dau[i]) > 0 {
			rows = append(rows, fmt.Sprintf("%.2f%%", float32(len(createChild[i])*100)/float32(len(v))))
		} else {
			rows = append(rows, "0.00%")
		}
		items = append(items, rows)
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("sys_user"), heads, items),
	}
}

func getSumAuth(gameId int32, startDate, endDate string, channel string, device string) []map[string]bool {
	var all map[string]bool
	retAll := []map[string]bool{}
	data.GetLogData(gameId, "RoomFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos == 0 {
			all = make(map[string]bool)
		} else if pos > 0 {
			if len(s) < 11 {
				logs.Error("getSumAuth data error!")
				return false
			}
			if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
				return true
			}
			if s[5] != "" {
				all[s[0]] = true
			}
			return true
		} else {
			retAll = append(retAll, all)
		}
		return true
	})
	return retAll
}
