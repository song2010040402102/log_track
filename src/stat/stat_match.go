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

type MatchInfo struct {
	sumPers int32
	mvpNum  int32
	sumCash int32
}

func MatchHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	matchInfo := getMatchInfo(gameId, startDate, endDate, channel, device)
	if len(matchInfo) == 0 {
		return nil
	}
	roomIds := make([]int32, 0, len(matchInfo))
	for k, _ := range matchInfo {
		roomIds = append(roomIds, k)
	}
	sort.Slice(roomIds, func(i, j int) bool { return roomIds[i] < roomIds[j] })
	heads := make([]string, 0, len(roomIds))
	for _, roomId := range roomIds {
		heads = append(heads, g_mapRoom[roomId])
	}
	childs := make([][]*TreeTable, 1)
	childs[0] = make([]*TreeTable, 0, len(roomIds))
	for _, roomId := range roomIds {
		heads := []string{"日期", "渠道", "设备", "红包总额(分)", "总人数", "赢红包人数", "胜率"}
		items := make([][]string, 0, len(matchInfo[roomId]))
		for i, v := range matchInfo[roomId] {
			rows := make([]string, 0, len(heads))
			rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
			rows = append(rows, channel)
			rows = append(rows, device)
			rows = append(rows, strconv.Itoa(int(v.sumCash)))
			rows = append(rows, strconv.Itoa(int(v.sumPers)))
			rows = append(rows, strconv.Itoa(int(v.mvpNum)))
			if v.sumPers > 0 {
				rows = append(rows, fmt.Sprintf("%.2f%%", float32(v.mvpNum)*100/float32(v.sumPers)))
			} else {
				rows = append(rows, "0.00%")
			}
			items = append(items, rows)
		}
		childs[0] = append(childs[0], NewTable("", heads, items))
	}
	return &ItemResult{
		TTable: NewTreeTable(fmt.Sprintf("比赛场红包数据统计(%s ~ %s)", startDate, endDate), heads, childs),
	}
}

func getMatchInfo(gameId int32, startDate, endDate string, channel string, device string) map[int32][]*MatchInfo {
	ret := make(map[int32][]*MatchInfo)
	days := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	data.GetLogData(gameId, "MatchResultFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		if len(s) < 16 {
			logs.Error("getMatchInfo data error!")
			return false
		}
		if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
			return true
		}

		id1, _ := strconv.ParseInt(s[13], 10, 32)
		id2, _ := strconv.ParseInt(s[12], 10, 32)
		if Is3PersRule(int32(id2)) {
			id1 = id1 * 4 / 3
		}
		id1 += 100
		if !GameHasRoom(gameId, int32(id1)) {
			return true
		}
		roomIds := [2]int32{0, int32(id1)}
		for _, roomId := range roomIds {
			if _, ok := ret[roomId]; !ok {
				data := make([]*MatchInfo, days)
				for i := 0; i < days; i++ {
					data[i] = &MatchInfo{}
				}
				ret[roomId] = data
			}
			ret[roomId][day].sumPers++
			if s[14] == "1" {
				ret[roomId][day].mvpNum++
				if s[13] == "4" {
					ret[roomId][day].sumCash += 600
				} else if s[13] == "16" {
					ret[roomId][day].sumCash += 1000
				} else if s[13] == "32" {
					ret[roomId][day].sumCash += 2000
				} else if s[13] == "3" {
					ret[roomId][day].sumCash += 400
				} else if s[13] == "12" {
					ret[roomId][day].sumCash += 800
				} else if s[13] == "24" {
					ret[roomId][day].sumCash += 1600
				}
			}
		}
		return true
	})
	return ret
}
