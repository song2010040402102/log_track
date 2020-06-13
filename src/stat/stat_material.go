package stat

import (
	"common"
	"data"
	"fmt"
	"github.com/astaxie/beego/logs"
	"sort"
	"strconv"
	"strings"
	"util"
)

func MaterialOutputHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	title := fmt.Sprintf("材料产出(%s ~ %s)", startDate, endDate)
	return materialHandler(gameId, LOG_RES_OUTPUT, startDate, endDate, channel, device, title)
}

func MaterialConsumeHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	title := fmt.Sprintf("材料消耗(%s ~ %s)", startDate, endDate)
	return materialHandler(gameId, LOG_RES_CONSUME, startDate, endDate, channel, device, title)
}

func materialHandler(gameId, fType int32, startDate, endDate, channel, device, title string) *ItemResult {
	data, sysKeys := getMaterialData(gameId, fType, startDate, endDate, channel, device)
	days := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	dateStr := make([]string, days)
	for i := 0; i < days; i++ {
		dateStr[i] = util.Ts2date(util.Date2ts(startDate) + int64(i*86400))
	}

	var sourceHeads []string
	childs := make([][]*TreeTable, 1)
	for _, k := range sysKeys {
		srcChild := make([][]*TreeTable, 1)
		sourceHeads = append(sourceHeads, g_sysType[k])
		var materialID []int32
		for mID := range data[k] {
			materialID = append(materialID, mID)
		}
		sort.Slice(materialID, func(i, j int) bool { return materialID[i] < materialID[j] })
		var mstrs []string
		for _, v := range materialID {
			mstrs = append(mstrs, strconv.FormatInt(int64(v), 10))
		}
		for _, mID := range materialID {
			var heads []string
			if fType == LOG_RES_OUTPUT {
				heads = []string{"日期", "渠道", "设备", "产出总额", "产出人次", "人均产出", "独立人数", "人均次数"}
			} else if fType == LOG_RES_CONSUME {
				heads = []string{"日期", "渠道", "设备", "消耗总额", "消耗人次", "人均消耗", "独立人数", "人均次数"}
			}
			var items [][]string
			for j := 0; j < len(data[k][mID]); j++ {
				info := data[k][mID][j]
				var avgTimes, timesPerPerson float64
				if info.personTime > 0 {
					avgTimes = float64(info.sum) / float64(info.personTime)
					timesPerPerson = float64(info.personTime) / float64(len(info.uniquePerson))
				}
				row := []string{
					dateStr[j],
					channel,
					device,
					strconv.FormatInt(int64(info.sum), 10),
					strconv.FormatInt(int64(info.personTime), 10),
					strconv.FormatFloat(avgTimes, 'f', 2, 64),
					strconv.FormatInt(int64(len(info.uniquePerson)), 10),
					strconv.FormatFloat(timesPerPerson, 'f', 2, 64),
				}
				items = append(items, row)
			}
			srcChild[0] = append(srcChild[0], NewTable("", heads, items))
		}
		childs[0] = append(childs[0], NewTreeTable("", mstrs, srcChild))
	}
	return &ItemResult{
		TTable: NewTreeTable(title, sourceHeads, childs),
	}
}

func getMaterialData(gameId, fType int32, startDate, endDate, channel, device string) (map[int32]map[int32][]*currencyInfo, []int32) {
	var sysKeys []int32
	result := make(map[int32]map[int32][]*currencyInfo)
	days := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	data.GetLogData(gameId, "MaterialFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		if len(s) < 17 {
			logs.Error("MaterialFlow invalid length", len(s))
			return false
		}
		if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
			return true
		}

		flowType, _ := strconv.Atoi(s[15])
		if int32(flowType) == fType {
			source, _ := strconv.Atoi(s[14])
			src := int32(source)
			valueDelta, _ := strconv.Atoi(s[13])
			id, _ := strconv.Atoi(strings.TrimSpace(s[16]))
			mID := int32(id)
			if _, ok := result[src]; !ok {
				result[src] = make(map[int32][]*currencyInfo)
				sysKeys = append(sysKeys, src)
			}
			if _, ok := result[src][mID]; !ok {
				result[src][mID] = make([]*currencyInfo, days)
				for j := 0; j < days; j++ {
					result[src][mID][j] = &currencyInfo{
						sum:          0,
						personTime:   0,
						uniquePerson: make(map[string]bool),
					}
				}
			}
			result[src][mID][day].sum += int32(valueDelta)
			result[src][mID][day].personTime++
			result[src][mID][day].uniquePerson[s[0]] = true
		}
		return true
	})
	sort.Slice(sysKeys, func(i, j int) bool { return sysKeys[i] < sysKeys[j] })
	return result, sysKeys
}
