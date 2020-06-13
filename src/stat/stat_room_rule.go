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

type RoomRuleInfo struct {
	sumRound int32
	mapUsers map[string]bool
}

func RoomRuleHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	roomIds := GetRoomIds(gameId)
	roomRule := getRoomRuleInfo(gameId, startDate, endDate, channel, device)
	ruleIds := getSortRuleIds(roomRule, 20)
	heads := make([]string, 0, len(ruleIds)+1)
	heads = append(heads, "")
	for _, ruleId := range ruleIds {
		heads = append(heads, g_mapRule[ruleId])
	}
	childs := make([][]*TreeTable, 0, len(roomIds))
	for _, roomId := range roomIds {
		childx := make([]*TreeTable, 0, len(ruleIds)+1)
		childx = append(childx, NewTable(g_mapRoom[roomId], nil, nil))
		for _, ruleId := range ruleIds {
			heads := []string{"日期", "渠道", "设备", "总场数", "总人数", "人均场数"}
			items := make([][]string, 0, len(roomRule[roomId][ruleId]))
			for i, v := range roomRule[roomId][ruleId] {
				rows := make([]string, 0, len(heads))
				if i == len(roomRule[roomId][ruleId])-1 {
					rows = append(rows, "汇总")
				} else {
					rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
				}
				rows = append(rows, channel)
				rows = append(rows, device)
				rows = append(rows, strconv.Itoa(int(v.sumRound)))
				rows = append(rows, strconv.Itoa(len(v.mapUsers)))
				if len(v.mapUsers) > 0 {
					rows = append(rows, strconv.FormatFloat(float64(v.sumRound)/float64(len(v.mapUsers)), 'f', 2, 64))
				} else {
					rows = append(rows, "0.00")
				}
				items = append(items, rows)
			}
			childx = append(childx, NewTable("", heads, items))
		}
		childs = append(childs, childx)
	}
	return &ItemResult{
		TTable: NewTreeTable(fmt.Sprintf("房间和玩法统计(%s ~ %s)", startDate, endDate), heads, childs),
	}
}

func getRoomRuleInfo(gameId int32, startDate, endDate string, channel string, device string) map[int32]map[int32][]*RoomRuleInfo {
	day := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	mapRoom := make(map[int32]map[int32][]*RoomRuleInfo)
	for room, _ := range g_mapRoom {
		if !GameHasRoom(gameId, room) {
			continue
		}
		mapRule := make(map[int32][]*RoomRuleInfo)
		for rule, _ := range g_mapRule {
			if !GameHasRule(gameId, rule) {
				continue
			}
			data := make([]*RoomRuleInfo, day+1)
			for i := 0; i < len(data); i++ {
				data[i] = &RoomRuleInfo{
					mapUsers: make(map[string]bool),
				}
			}
			mapRule[rule] = data
		}
		mapRoom[room] = mapRule
	}
	handler := func(k int, d int32, s []string) bool {
		if len(s) < 13 {
			logs.Error("getRoomRuleInfo data error!")
			return false
		}
		if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
			return true
		}
		id1, id2 := int64(0), int64(0)
		if k == 0 {
			id1, _ = strconv.ParseInt(s[12], 10, 32)
			id2, _ = strconv.ParseInt(s[11], 10, 32)
		} else if k == 1 {
			id1, _ = strconv.ParseInt(s[13], 10, 32)
			id2, _ = strconv.ParseInt(s[12], 10, 32)
			if Is3PersRule(int32(id2)) {
				id1 = id1 * 4 / 3
			}
			id1 += 100
		} else {
			id1, _ = strconv.ParseInt(s[11], 10, 32)
			id2, _ = strconv.ParseInt(s[12], 10, 32)
		}
		roomId, ruleId := int32(id1), int32(id2)
		roomIds := []int32{roomId, 0}
		ruleIds := []int32{ruleId, 0}
		days := []int32{d, int32(day)}
		for _, rmId := range roomIds {
			for _, rlId := range ruleIds {
				for _, day := range days {
					if _, ok1 := mapRoom[rmId]; ok1 {
						if _, ok2 := mapRoom[rmId][rlId]; ok2 {
							mapRoom[rmId][rlId][day].sumRound++
							mapRoom[rmId][rlId][day].mapUsers[s[0]] = true
						}
					}
				}
			}
		}
		return true
	}
	err := data.GetLogData(gameId, "RoomFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		return handler(0, day, s)
	})
	if err != nil {
		return nil
	}
	err = data.GetLogData(gameId, "MatchFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		return handler(1, day, s)
	})
	if err != nil {
		return nil
	}
	err = data.GetLogData(gameId, "GrandPrixFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		return handler(2, day, s)
	})
	if err != nil {
		return nil
	}
	return mapRoom
}

func getSortRuleIds(roomRule map[int32]map[int32][]*RoomRuleInfo, n int32) []int32 {
	mRuleNum := make(map[int32]int32)
	for _, room := range roomRule {
		for r, rule := range room {
			if len(rule) == 0 {
				continue
			}
			mRuleNum[r] += rule[len(rule)-1].sumRound
		}
	}
	ruleIds := make([]int32, 0, len(mRuleNum))
	for rId, _ := range mRuleNum {
		ruleIds = append(ruleIds, rId)
	}
	sort.Slice(ruleIds, func(i, j int) bool { return mRuleNum[ruleIds[i]] > mRuleNum[ruleIds[j]] })
	if int32(len(ruleIds)) > n {
		ruleIds = ruleIds[:n]
	}
	return ruleIds
}
