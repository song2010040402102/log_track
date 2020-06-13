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

type DMPInfo struct {
	winSum  int32
	winNum  int32
	loseSum int32
	loseNum int32
	zeroNum int32
}

func DDZMergePlayHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	dmpInfo := getDMPInfo(gameId, startDate, endDate, channel, device)
	if len(dmpInfo) == 0 {
		return nil
	}
	levels := make([]int32, 0, len(dmpInfo))
	for k, _ := range dmpInfo {
		levels = append(levels, k)
	}
	sort.Slice(levels, func(i, j int) bool { return levels[i] < levels[j] })
	ruleIds := GetRuleIds(gameId)
	heads := make([]string, 0, len(ruleIds)+1)
	heads = append(heads, "")
	for _, ruleId := range ruleIds {
		heads = append(heads, g_mapRule[ruleId])
	}
	childs := make([][]*TreeTable, 0, len(levels))
	for _, level := range levels {
		childx := make([]*TreeTable, 0, len(ruleIds)+1)
		childx = append(childx, NewTable(fmt.Sprintf("%d档", level), nil, nil))
		for _, ruleId := range ruleIds {
			heads := []string{"日期", "渠道", "设备", "赢总额", "赢次数", "输总额", "输次数", "归零次数"}
			items := make([][]string, 0, len(dmpInfo[level][ruleId]))
			for i, v := range dmpInfo[level][ruleId] {
				rows := make([]string, 0, len(heads))
				rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
				rows = append(rows, channel)
				rows = append(rows, device)
				rows = append(rows, strconv.Itoa(int(v.winSum)))
				rows = append(rows, strconv.Itoa(int(v.winNum)))
				rows = append(rows, strconv.Itoa(int(v.loseSum)))
				rows = append(rows, strconv.Itoa(int(v.loseNum)))
				rows = append(rows, strconv.Itoa(int(v.zeroNum)))
				items = append(items, rows)
			}
			childx = append(childx, NewTable("", heads, items))
		}
		childs = append(childs, childx)
	}
	return &ItemResult{
		TTable: NewTreeTable(fmt.Sprintf("斗地主强控数据统计(%s ~ %s)", startDate, endDate), heads, childs),
	}
}

func getDMPInfo(gameId int32, startDate, endDate string, channel string, device string) map[int32]map[int32][]*DMPInfo {
	days := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	ret := make(map[int32]map[int32][]*DMPInfo)
	err := data.GetLogData(gameId, "RoomMergePlayFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		if len(s) < 18 {
			logs.Error("getDMPInfo data error!")
			return false
		}
		if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
			return true
		}
		l, _ := strconv.ParseInt(s[13], 10, 32)
		level := int32(l)
		if level <= 0 {
			return true
		}
		rId, _ := strconv.ParseInt(s[11], 10, 32)
		if !GameHasRule(gameId, int32(rId)) {
			return true
		}
		if _, ok := ret[level]; !ok {
			m := make(map[int32][]*DMPInfo)
			for rule, _ := range g_mapRule {
				if !GameHasRule(gameId, rule) {
					continue
				}
				data := make([]*DMPInfo, days)
				for i := 0; i < days; i++ {
					data[i] = &DMPInfo{}
				}
				m[rule] = data
			}
			ret[level] = m
		}
		cash, _ := strconv.ParseInt(s[16], 10, 32)
		ruleIds := []int32{0, int32(rId)}
		for _, ruleId := range ruleIds {
			if _, ok := ret[level][ruleId]; !ok {
				continue
			}
			if s[17] == "1" {
				ret[level][ruleId][day].winSum += int32(cash)
				ret[level][ruleId][day].winNum++
			} else {
				ret[level][ruleId][day].loseSum -= int32(cash)
				ret[level][ruleId][day].loseNum++
				if s[15] == "0" {
					ret[level][ruleId][day].zeroNum++
				}
			}
		}
		return true
	})
	if err != nil {
		return nil
	}
	return ret
}
