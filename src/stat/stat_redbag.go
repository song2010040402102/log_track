package stat

import (
	"common"
	"data"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
	"strings"
	"util"
)

type redBagInfo struct {
	sum          int32           // 总额
	personTime   int32           // 人次
	uniquePerson map[string]bool // 独立用户
}

// RedBagOutPutHandler 提供红包产出的数据
func RedBagOutPutHandler(gameId int32, startDate, endDate, channel string, device string) *ItemResult {
	days := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	datestr := make([]string, days)
	result := make([]map[int]*redBagInfo, days)

	for i := 0; i < days; i++ {
		datestr[i] = util.Ts2date(util.Date2ts(startDate) + int64(i*86400))
		result[i] = make(map[int]*redBagInfo)
	}
	totalTimes := make([]int32, days)
	totalSum := make([]int32, days)
	totalUniquePerson := make([]int, days)
	overallUnique := make(map[string]bool)

	var output map[int]*redBagInfo
	var dailyUnique map[string]bool
	err := data.GetLogData(gameId, "RedBagFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos == 0 {
			output = result[day]
			dailyUnique = make(map[string]bool)
		} else if pos > 0 {
			if len(s) < 14 {
				logs.Error("RedBagData output invalid length", len(s))
				return false
			}
			if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
				return true
			}

			ch, _ := strconv.Atoi(s[11])
			outPutVal, _ := strconv.Atoi(s[10])
			if _, ok := output[ch]; !ok {
				for j := 0; j < days; j++ {
					result[j][ch] = &redBagInfo{
						sum:          0,
						personTime:   0,
						uniquePerson: make(map[string]bool),
					}
				}
			}
			output[ch].sum += int32(outPutVal)
			output[ch].personTime++
			output[ch].uniquePerson[s[0]] = true
			dailyUnique[s[0]] = true
			overallUnique[s[0]] = true
		} else {
			totalUniquePerson[day] = len(dailyUnique)
		}
		return true
	})
	if err != nil {
		return nil
	}

	// systype - data
	itemsMap := make(map[int32][][]string)
	for i, output := range result {
		for j := range g_sysType {
			if op, ok := output[int(j)]; ok {
				totalSum[i] += op.sum
				totalTimes[i] += op.personTime
				var outputPerTime, timesPerPerson float64
				if op.personTime > 0 {
					outputPerTime = float64(op.sum) / float64(op.personTime)
					timesPerPerson = float64(op.personTime) / float64(len(op.uniquePerson))
				}
				row := []string{
					datestr[i],
					channel,
					device,
					strconv.FormatInt(int64(op.sum), 10),
					strconv.FormatInt(int64(op.personTime), 10),
					strconv.FormatFloat(outputPerTime, 'f', 2, 64),
					strconv.FormatInt(int64(len(op.uniquePerson)), 10),
					strconv.FormatFloat(timesPerPerson, 'f', 2, 64),
				}
				itemsMap[j] = append(itemsMap[j], row)
			}
		}
	}

	var tables []*TreeTable
	var overallHeads []string
	// 专门汇总的表
	var totalItem [][]string
	var overallSum, overallPersonTimes int32
	var perTime, perPerson float64
	for i := 0; i < days; i++ {
		overallSum += totalSum[i]
		overallPersonTimes += totalTimes[i]
		var totalopPerTime, totaltiPerPerson float64
		if totalTimes[i] > 0 {
			totalopPerTime = float64(totalSum[i]) / float64(totalTimes[i])
			totaltiPerPerson = float64(totalTimes[i]) / float64(totalUniquePerson[i])
		}
		totalItem = append(totalItem, []string{
			datestr[i],
			channel,
			device,
			strconv.FormatInt(int64(totalSum[i]), 10),
			strconv.FormatInt(int64(totalTimes[i]), 10),
			strconv.FormatFloat(totalopPerTime, 'f', 2, 64),
			strconv.FormatInt(int64(totalUniquePerson[i]), 10),
			strconv.FormatFloat(totaltiPerPerson, 'f', 2, 64),
		})
	}

	if overallPersonTimes > 0 {
		perTime = float64(overallSum) / float64(overallPersonTimes)
		perPerson = float64(overallPersonTimes) / float64(len(overallUnique))
	}
	totalItem = append(totalItem, []string{
		"汇总",
		channel,
		device,
		strconv.FormatInt(int64(overallSum), 10),
		strconv.FormatInt(int64(overallPersonTimes), 10),
		strconv.FormatFloat(perTime, 'f', 2, 64),
		strconv.FormatInt(int64(len(overallUnique)), 10),
		strconv.FormatFloat(perPerson, 'f', 2, 64),
	})
	itemsMap[0] = totalItem

	keys := sortedSystemKeys()
	for _, v := range keys {
		overallHeads = append(overallHeads, g_sysType[v])
		if it, ok := itemsMap[v]; ok {
			heads := []string{"日期", "渠道", "设备", "产出总额", "产出人次", "平均产出", "独立人数", "人均次数"}
			t := NewTable("", heads, it)
			tables = append(tables, t)
		} else {
			tables = append(tables, &TreeTable{})
		}
	}
	return &ItemResult{
		TTable: NewTreeTable(fmt.Sprintf("红包产出(%s ~ %s)", startDate, endDate), overallHeads, [][]*TreeTable{tables}),
	}
}

func RedBagConsumeHandler(gameId int32, startDate, endDate, channel string, device string) *ItemResult {
	days := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	datestr := make([]string, days)
	stats := make([]map[int]*redBagInfo, days)
	for i := 0; i < days; i++ {
		datestr[i] = util.Ts2date(util.Date2ts(startDate) + int64(i*86400))
		stats[i] = make(map[int]*redBagInfo)
	}
	totalTimes := make([]int32, days)
	totalSum := make([]int32, days)
	totalUniquePerson := make([]int, days)
	overallUnique := make(map[string]bool)

	var cost map[int]*redBagInfo
	var dailyUnique map[string]bool
	err := data.GetLogData(gameId, "RedbagConsumeFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos == 0 {
			cost = stats[day]
			dailyUnique = make(map[string]bool)
		} else if pos > 0 {
			if len(s) < 16 {
				logs.Error("Red Bag consume data length error", len(s))
				return false
			}
			if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
				return true
			}
			ch, _ := strconv.Atoi(s[14])
			flowType := strings.TrimSpace(s[15])
			if flowType == "2" {
				costDelta, _ := strconv.Atoi(s[13])
				if _, ok := cost[ch]; !ok {
					for j := 0; j < days; j++ {
						stats[j][ch] = &redBagInfo{
							sum:          0,
							personTime:   0,
							uniquePerson: make(map[string]bool),
						}
					}
				}
				cost[ch].sum += int32(costDelta)
				cost[ch].personTime++
				cost[ch].uniquePerson[s[0]] = true
				dailyUnique[s[0]] = true
				overallUnique[s[0]] = true
			}
			return true
		} else {
			totalUniquePerson[day] += len(dailyUnique)
		}
		return true
	})
	if err != nil {
		return nil
	}

	itemsMap := make(map[int32][][]string)
	for i, cost := range stats {
		for j := range g_sysType {
			if co, ok := cost[int(j)]; ok {
				totalSum[i] += co.sum
				totalTimes[i] += co.personTime
				var costPerTime, timesPerPerson float64
				if co.personTime > 0 {
					costPerTime = float64(co.sum) / float64(co.personTime)
				}
				if len(co.uniquePerson) > 0 {
					timesPerPerson = float64(co.personTime) / float64(len(co.uniquePerson))
				}
				row := []string{
					datestr[i],
					channel,
					device,
					strconv.FormatInt(int64(co.sum), 10),
					strconv.FormatInt(int64(co.personTime), 10),
					strconv.FormatFloat(costPerTime, 'f', 2, 64),
					strconv.FormatInt(int64(len(co.uniquePerson)), 10),
					strconv.FormatFloat(timesPerPerson, 'f', 2, 64),
				}
				itemsMap[j] = append(itemsMap[j], row)
			}
		}
	}

	var tables []*TreeTable
	var overallHeads []string
	var totalItem [][]string
	var overallSum, overallPersonTimes int32
	var perTime, perPerson float64
	for i := 0; i < days; i++ {
		overallSum += totalSum[i]
		overallPersonTimes += totalTimes[i]
		var totalcoPerTime, totaltiPerPerson float64
		if totalTimes[i] > 0 {
			totalcoPerTime = float64(totalSum[i]) / float64(totalTimes[i])
			totaltiPerPerson = float64(totalTimes[i]) / float64(totalUniquePerson[i])
		}
		totalItem = append(totalItem, []string{
			datestr[i],
			channel,
			device,
			strconv.FormatInt(int64(totalSum[i]), 10),
			strconv.FormatInt(int64(totalTimes[i]), 10),
			strconv.FormatFloat(totalcoPerTime, 'f', 2, 64),
			strconv.FormatInt(int64(totalUniquePerson[i]), 10),
			strconv.FormatFloat(totaltiPerPerson, 'f', 2, 64),
		})
	}
	if overallPersonTimes > 0 {
		perTime = float64(overallSum) / float64(overallPersonTimes)
		perPerson = float64(overallPersonTimes) / float64(len(overallUnique))
	}
	totalItem = append(totalItem, []string{
		"汇总",
		channel,
		device,
		strconv.FormatInt(int64(overallSum), 10),
		strconv.FormatInt(int64(overallPersonTimes), 10),
		strconv.FormatFloat(perTime, 'f', 2, 64),
		strconv.FormatInt(int64(len(overallUnique)), 10),
		strconv.FormatFloat(perPerson, 'f', 2, 64),
	})
	itemsMap[0] = totalItem

	keys := sortedSystemKeys()
	for _, v := range keys {
		overallHeads = append(overallHeads, g_sysType[v])
		if it, ok := itemsMap[v]; ok {
			heads := []string{"日期", "渠道", "设备", "消耗总额", "消耗人次", "平均消耗", "独立人数", "人均次数"}
			t := NewTable("", heads, it)
			tables = append(tables, t)
		} else {
			tables = append(tables, &TreeTable{})
		}
	}
	return &ItemResult{
		TTable: NewTreeTable(fmt.Sprintf("红包消耗(%s ~ %s)", startDate, endDate), overallHeads, [][]*TreeTable{tables}),
	}
}
