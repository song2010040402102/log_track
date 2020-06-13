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

func GrandPrixHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	grandInfo := getGrandPrixInfo(gameId, startDate, endDate, channel, device)
	childs := make([][]*TreeTable, 1)
	childs[0] = make([]*TreeTable, 0, len(grandInfo))
	for _, mGrand := range grandInfo {
		roomIds := make([]int32, 0, len(mGrand))
		for k, _ := range mGrand {
			roomIds = append(roomIds, k)
		}
		sort.Slice(roomIds, func(i, j int) bool { return roomIds[i] < roomIds[j] })
		gHeads := make([]string, 0, len(mGrand))
		for _, roomId := range roomIds {
			gHeads = append(gHeads, g_mapRoom[roomId])
		}
		gChilds := make([][]*TreeTable, 1)
		gChilds[0] = make([]*TreeTable, 0, len(mGrand))
		for _, roomId := range roomIds {
			heads := []string{"日期", "渠道", "设备", "总金额(分)", "总人数", "TOP1人数", "胜率"}
			items := make([][]string, 0, len(mGrand[roomId]))
			for i, v := range mGrand[roomId] {
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
			gChilds[0] = append(gChilds[0], NewTable("", heads, items))
		}
		childs[0] = append(childs[0], NewTreeTable("", gHeads, gChilds))
	}
	return &ItemResult{
		TTable: NewTreeTable(fmt.Sprintf("大奖赛数据统计(%s ~ %s)", startDate, endDate), []string{"红包赛", "话费赛"}, childs),
	}
}

func getGrandPrixInfo(gameId int32, startDate, endDate string, channel string, device string) [2]map[int32][]*MatchInfo {
	ret := [2]map[int32][]*MatchInfo{}
	ret[0] = make(map[int32][]*MatchInfo)
	ret[1] = make(map[int32][]*MatchInfo)
	days := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	data.GetLogData(gameId, "GrandPrixResultFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		if len(s) < 15 {
			logs.Error("getGrandPrixInfo data error!")
			return false
		}
		if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
			return true
		}

		id, _ := strconv.ParseInt(s[11], 10, 32)
		if !GameHasRoom(gameId, int32(id)) {
			return true
		}
		index := 0
		if id >= 3001 && id <= 3003 {
			index = 1
		}
		roomIds := [2]int32{0, int32(id)}
		for _, roomId := range roomIds {
			if _, ok := ret[index][roomId]; !ok {
				data := make([]*MatchInfo, days)
				for i := 0; i < days; i++ {
					data[i] = &MatchInfo{}
				}
				ret[index][roomId] = data
			}
			ret[index][roomId][day].sumPers++
			if s[14] == "1" {
				ret[index][roomId][day].mvpNum++
				if id == 2001 {
					ret[index][roomId][day].sumCash += 300
				} else if id == 2002 {
					ret[index][roomId][day].sumCash += 1000
				} else if id == 2003 {
					ret[index][roomId][day].sumCash += 2000
				} else if id == 3001 {
					ret[index][roomId][day].sumCash += 100
				} else if id == 3002 {
					ret[index][roomId][day].sumCash += 800
				} else if id == 3003 {
					ret[index][roomId][day].sumCash += 1000
				}
			}
		}
		return true
	})
	return ret
}
