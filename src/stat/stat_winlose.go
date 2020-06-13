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

const GOLD_LOWER_BOUND int32 = 100

type WinLoseInfo struct {
	winSum    int64
	winNum    int32
	winUsers  map[string]bool
	loseSum   int64
	loseNum   int32
	loseUsers map[string]bool
}

func GoldWinLoseHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	return winLoseHandler(gameId, startDate, endDate, channel, device, ROOM_GOLD, 0)
}

func GoldWinLoseBigHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	return winLoseHandler(gameId, startDate, endDate, channel, device, ROOM_GOLD, GOLD_LOWER_BOUND)
}

func PlaceWinLoseHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	return winLoseHandler(gameId, startDate, endDate, channel, device, ROOM_PLACE, 0)
}

func winLoseHandler(gameId int32, startDate, endDate string, channel string, device string, roomId int32, lr int32) *ItemResult {
	wlInfo := getWinLoseInfo(gameId, startDate, endDate, channel, device, roomId, lr)
	if len(wlInfo) == 0 {
		return nil
	}
	levels := make([]int32, 0, len(wlInfo))
	for k, _ := range wlInfo {
		levels = append(levels, k)
	}
	sort.Slice(levels, func(i, j int) bool { return levels[i] < levels[j] })
	ruleIds := GetRuleIds(gameId)
	heads := make([]string, 0, len(ruleIds)+1)
	if roomId == ROOM_GOLD {
		heads = append(heads, "")
	}
	for _, ruleId := range ruleIds {
		heads = append(heads, g_mapRule[ruleId])
	}
	childs := make([][]*TreeTable, 0, len(levels))
	for _, level := range levels {
		childx := make([]*TreeTable, 0, len(ruleIds)+1)
		if roomId == ROOM_GOLD {
			childx = append(childx, NewTable(fmt.Sprintf("%d档", level), nil, nil))
		}
		for _, ruleId := range ruleIds {
			heads := []string{"日期", "渠道", "设备", "赢总额", "赢次数", "赢人数", "场均赢", "人均赢", "输总额", "输次数", "输人数", "场均输", "人均输", "平均胜率", "系统支出"}
			items := make([][]string, 0, len(wlInfo[level][ruleId]))
			for i, v := range wlInfo[level][ruleId] {
				rows := make([]string, 0, len(heads))
				if i == len(wlInfo[level][ruleId])-1 {
					rows = append(rows, "汇总")
				} else {
					rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
				}
				rows = append(rows, channel)
				rows = append(rows, device)
				rows = append(rows, strconv.FormatInt(v.winSum, 10))
				rows = append(rows, strconv.Itoa(int(v.winNum)))
				rows = append(rows, strconv.Itoa(len(v.winUsers)))
				if v.winNum > 0 {
					rows = append(rows, strconv.FormatFloat(float64(v.winSum)/float64(v.winNum), 'f', 2, 64))
				} else {
					rows = append(rows, "0.00")
				}
				if len(v.winUsers) > 0 {
					rows = append(rows, strconv.FormatFloat(float64(v.winSum)/float64(len(v.winUsers)), 'f', 2, 64))
				} else {
					rows = append(rows, "0.00")
				}
				rows = append(rows, strconv.FormatInt(v.loseSum, 10))
				rows = append(rows, strconv.Itoa(int(v.loseNum)))
				rows = append(rows, strconv.Itoa(len(v.loseUsers)))
				if v.loseNum > 0 {
					rows = append(rows, strconv.FormatFloat(float64(v.loseSum)/float64(v.loseNum), 'f', 2, 64))
				} else {
					rows = append(rows, "0.00")
				}
				if len(v.loseUsers) > 0 {
					rows = append(rows, strconv.FormatFloat(float64(v.loseSum)/float64(len(v.loseUsers)), 'f', 2, 64))
				} else {
					rows = append(rows, "0.00")
				}
				if sum := v.winNum + v.loseNum; sum > 0 {
					rows = append(rows, fmt.Sprintf("%.2f%%", float32(v.winNum)*100/float32(sum)))
				} else {
					rows = append(rows, "0.00%")
				}
				rows = append(rows, strconv.FormatInt(v.winSum-v.loseSum, 10))
				items = append(items, rows)
			}
			childx = append(childx, NewTable("", heads, items))
		}
		childs = append(childs, childx)
	}
	title := ""
	if roomId == ROOM_GOLD {
		if lr == 0 {
			title = "金币场不同玩法不同档位胜率相关统计"
		} else {
			title = fmt.Sprintf("金币场不同玩法不同档位胜率(赢%d倍)相关统计", lr)
		}
	} else if roomId == ROOM_PLACE {
		if lr == 0 {
			title = "排位赛不同玩法不同档位胜率相关统计"
		} else {
			title = fmt.Sprintf("排位赛不同玩法不同档位胜率(赢%d倍)相关统计", lr)
		}
	}
	return &ItemResult{
		TTable: NewTreeTable(fmt.Sprintf("%s(%s ~ %s)", title, startDate, endDate), heads, childs),
	}
}

func getWinLoseInfo(gameId int32, startDate, endDate string, channel string, device string, roomId int32, lr int32) map[int32]map[int32][]*WinLoseInfo {
	ret := make(map[int32]map[int32][]*WinLoseInfo)
	days := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	var flow string
	if roomId == ROOM_GOLD {
		flow = "GoldFlow"
	} else if roomId == ROOM_PLACE {
		flow = "StarIntegalFlow"
	} else {
		logs.Error("getWinLoseInfo, invalid roomId!")
		return ret
	}
	data.GetLogData(gameId, flow, startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		if len(s) < 17 {
			logs.Error("getWinLoseInfo data error!")
			return false
		}
		if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
			return true
		}
		filter := ""
		if roomId == ROOM_GOLD {
			filter = "7"
		} else {
			filter = "35"
		}
		if s[14] != filter {
			return true
		}
		level := int32(0)
		if roomId == ROOM_GOLD {
			l, _ := strconv.ParseInt(s[16], 10, 32)
			level = int32(l)
			if level <= 0 {
				return true
			}
		}
		rId, _ := strconv.ParseInt(s[10], 10, 32)
		if !GameHasRule(gameId, int32(rId)) {
			return true
		}

		if _, ok := ret[level]; !ok {
			m := make(map[int32][]*WinLoseInfo)
			for rule, _ := range g_mapRule {
				if !GameHasRule(gameId, rule) {
					continue
				}
				data := make([]*WinLoseInfo, days+1)
				for i := 0; i <= days; i++ {
					data[i] = &WinLoseInfo{
						winUsers:  make(map[string]bool),
						loseUsers: make(map[string]bool),
					}
				}
				m[rule] = data
			}
			ret[level] = m
		}
		num, _ := strconv.ParseInt(s[13], 10, 32)
		if int32(num) <= level*lr {
			return true
		}
		ruleIds := []int32{0, int32(rId)}
		for _, ruleId := range ruleIds {
			if _, ok := ret[level][ruleId]; !ok {
				continue
			}
			si := []int{int(day), days}
			for _, i := range si {
				if s[15] == "1" {
					ret[level][ruleId][i].winSum += num
					ret[level][ruleId][i].winNum++
					ret[level][ruleId][i].winUsers[s[0]] = true
				} else {
					ret[level][ruleId][i].loseSum += num
					ret[level][ruleId][i].loseNum++
					ret[level][ruleId][i].loseUsers[s[0]] = true
				}
			}
		}
		return true
	})
	return ret
}
