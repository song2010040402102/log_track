package stat

import (
	"common"
	"data"
	"fmt"
	"github.com/astaxie/beego/logs"
	"sort"
	"strconv"
	"util"
)

type ShareInfo struct {
	sum      int32
	newUsers map[string]bool
}

func SharePicHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	shareInfo := getShareInfo(gameId, startDate, endDate, channel, device)
	if len(shareInfo) == 0 {
		return nil
	}
	appIds := make([]int32, 0, len(shareInfo))
	for k, _ := range shareInfo {
		appIds = append(appIds, k)
	}
	sort.Slice(appIds, func(i, j int) bool { return appIds[i] < appIds[j] })
	heads := make([]string, 0, len(appIds))
	heads = append(heads, "all")
	for i := 1; i < len(appIds); i++ {
		heads = append(heads, strconv.Itoa(int(appIds[i])))
	}
	childs := make([][]*TreeTable, 1)
	childs[0] = make([]*TreeTable, 0, len(appIds))
	for _, appId := range appIds {
		heads := []string{"日期", "渠道", "设备", "总点击量", "新用户数"}
		items := make([][]string, 0, len(shareInfo[appId]))
		for i, v := range shareInfo[appId] {
			rows := make([]string, 0, len(heads))
			rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
			rows = append(rows, channel)
			rows = append(rows, device)
			rows = append(rows, strconv.Itoa(int(v.sum)))
			rows = append(rows, strconv.Itoa(len(v.newUsers)))
			items = append(items, rows)
		}
		childs[0] = append(childs[0], NewTable("", heads, items))
	}
	return &ItemResult{
		TTable: NewTreeTable(fmt.Sprintf("各分享图流入新用户数据统计(%s ~ %s)", startDate, endDate), heads, childs),
	}
}

func getShareInfo(gameId int32, startDate, endDate string, channel string, device string) map[int32][]*ShareInfo {
	ret := make(map[int32][]*ShareInfo)
	days := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	data.GetLogData(gameId, "WeChatSharePicFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		if len(s) < 13 {
			logs.Error("getShareInfo data error!")
			return false
		}
		if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
			return true
		}
		id, _ := strconv.ParseInt(s[10], 10, 32)
		appIds := []int32{0, int32(id)}
		for _, v := range appIds {
			if _, ok := ret[v]; !ok {
				data := make([]*ShareInfo, days)
				for i := 0; i < days; i++ {
					data[i] = &ShareInfo{
						newUsers: make(map[string]bool),
					}
				}
				ret[v] = data
			}
			ret[v][day].sum++
			if s[12] != "0" {
				ret[v][day].newUsers[s[0]] = true
			}
		}
		return true
	})
	return ret
}
