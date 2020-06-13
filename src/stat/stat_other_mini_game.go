package stat

import (
	"common"
	"data"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
	"util"
)

type OtherMiniGame struct {
	sum      int32
	mapUsers map[string]bool
}

func OtherMiniGameHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	miniInfo := getOtherMiniGame(gameId, startDate, endDate, channel, device)
	if len(miniInfo) == 0 {
		return nil
	}
	appIds := make([]string, 0, len(miniInfo))
	appIds = append(appIds, "all")
	for k, _ := range miniInfo {
		if k == "all" {
			continue
		}
		appIds = append(appIds, k)
	}
	childs := make([][]*TreeTable, 1)
	childs[0] = make([]*TreeTable, 0, len(appIds))
	for _, appId := range appIds {
		heads := []string{"日期", "渠道", "设备", "小游戏跳转次数", "人数", "复跳率"}
		items := make([][]string, 0, len(miniInfo[appId]))
		for i, v := range miniInfo[appId] {
			rows := make([]string, 0, len(heads))
			rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
			rows = append(rows, channel)
			rows = append(rows, device)
			rows = append(rows, strconv.Itoa(int(v.sum)))
			rows = append(rows, strconv.Itoa(len(v.mapUsers)))
			if len(v.mapUsers) > 0 {
				rows = append(rows, strconv.FormatFloat(float64(v.sum)/float64(len(v.mapUsers)), 'f', 2, 64))
			} else {
				rows = append(rows, "0.00")
			}
			items = append(items, rows)
		}
		childs[0] = append(childs[0], NewTable("", heads, items))
	}
	return &ItemResult{
		TTable: NewTreeTable(fmt.Sprintf("玩其他小游戏数据统计(%s ~ %s)", startDate, endDate), appIds, childs),
	}
}

func getOtherMiniGame(gameId int32, startDate, endDate string, channel string, device string) map[string][]*OtherMiniGame {
	ret := make(map[string][]*OtherMiniGame)
	days := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	data.GetLogData(gameId, "AdToMinGameFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		if len(s) < 11 {
			logs.Error("getOtherMiniGame data error!")
			return false
		}
		if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
			return true
		}
		appIds := []string{"all", s[10]}
		for _, v := range appIds {
			if _, ok := ret[v]; !ok {
				data := make([]*OtherMiniGame, days)
				for i := 0; i < days; i++ {
					data[i] = &OtherMiniGame{
						mapUsers: make(map[string]bool),
					}
				}
				ret[v] = data
			}
			ret[v][day].sum++
			ret[v][day].mapUsers[s[0]] = true
		}
		return true
	})
	return ret
}
