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

// 纸巾机积分,金币，钻石，报名券，段位分，大转盘通用结构体
type currencyInfo struct {
	sum          int32
	personTime   int32
	uniquePerson map[string]bool
}

// ZJJIntegalOutputHandler 纸巾机积分产出
func ZJJIntegalOutputHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	title := fmt.Sprintf("纸巾机积分产出(%s ~ %s)", startDate, endDate)
	return currencyHandler(gameId, LOG_RES_OUTPUT, startDate, endDate, channel, device, "ZjjIntegalFlow", title)
}

// ZJJIntegalConsumeHandler 纸巾机积分消耗
func ZJJIntegalConsumeHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	title := fmt.Sprintf("纸巾机积分消耗(%s ~ %s)", startDate, endDate)
	return currencyHandler(gameId, LOG_RES_CONSUME, startDate, endDate, channel, device, "ZjjIntegalFlow", title)
}

func currencyHandler(gameId, fType int32, startDate, endDate, channel, device, flowName, title string) *ItemResult {
	data, sysKeys := getCurrencyData(gameId, fType, startDate, endDate, channel, device, flowName)
	days := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	dateStr := make([]string, days)
	for i := 0; i < days; i++ {
		dateStr[i] = util.Ts2date(util.Date2ts(startDate) + int64(i*86400))
	}
	var sourceHeads []string
	var tables []*TreeTable
	for _, k := range sortedSystemKeys() {
		sourceHeads = append(sourceHeads, g_sysType[k])
		if util.InSlice(sysKeys, k) {
			var heads []string
			if fType == LOG_RES_OUTPUT {
				heads = []string{"日期", "渠道", "设备", "产出总额", "产出人次", "人均产出", "独立人数", "人均次数"}
			} else if fType == LOG_RES_CONSUME {
				heads = []string{"日期", "渠道", "设备", "消耗总额", "消耗人次", "人均消耗", "独立人数", "人均次数"}
			}
			var items [][]string
			for i := 0; i < len(data[k]); i++ {
				curInfo := data[k][i]
				var avgTimes, timesPerPerson float64
				if curInfo.personTime > 0 {
					avgTimes = float64(curInfo.sum) / float64(curInfo.personTime)
					timesPerPerson = float64(curInfo.personTime) / float64(len(curInfo.uniquePerson))
				}
				row := []string{
					dateStr[i],
					channel,
					device,
					strconv.FormatInt(int64(curInfo.sum), 10),
					strconv.FormatInt(int64(curInfo.personTime), 10),
					strconv.FormatFloat(avgTimes, 'f', 2, 64),
					strconv.FormatInt(int64(len(curInfo.uniquePerson)), 10),
					strconv.FormatFloat(timesPerPerson, 'f', 2, 64),
				}
				items = append(items, row)
			}
			tables = append(tables, NewTable("", heads, items))
		} else {
			tables = append(tables, &TreeTable{})
		}
	}
	return &ItemResult{
		TTable: NewTreeTable(title, sourceHeads, [][]*TreeTable{tables}),
	}
}

func getCurrencyData(gameId, fType int32, startDate, endDate, channel, device, flowName string) (map[int32][]*currencyInfo, []int32) {
	days := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	logLength := map[string]int32{
		"EnrollVoucherFlow": 16,
		"StarIntegalFlow":   18,
		"GoldFlow":          18,
		"DiamondFlow":       18,
		"WheelFlow":         16,
		"ZjjIntegalFlow":    18,
	}
	result := make(map[int32][]*currencyInfo)
	result[0] = make([]*currencyInfo, days)
	for i := 0; i < days; i++ {
		result[0][i] = &currencyInfo{
			sum:          0,
			personTime:   0,
			uniquePerson: make(map[string]bool),
		}
	}

	sysKeys := []int32{0}
	data.GetLogData(gameId, flowName, startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		if _, ok := logLength[flowName]; !ok {
			logs.Error("currency type does not exist ", flowName)
			return false
		}
		if len(s) < int(logLength[flowName]) {
			logs.Error("currency invalid length", flowName, len(s))
			return false
		}
		if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
			return true
		}
		source, _ := strconv.Atoi(s[14])
		src := int32(source)
		flowType, _ := strconv.Atoi(s[15])
		valueDelta, _ := strconv.Atoi(s[13])
		if int32(flowType) == fType {
			if _, ok := result[src]; !ok {
				result[src] = make([]*currencyInfo, days)
				for i := 0; i < days; i++ {
					result[src][i] = &currencyInfo{
						sum:          0,
						personTime:   0,
						uniquePerson: make(map[string]bool),
					}
				}
				sysKeys = append(sysKeys, src)
			}
			result[src][day].personTime++
			result[src][day].sum += int32(valueDelta)
			result[src][day].uniquePerson[s[0]] = true
			result[0][day].personTime++
			result[0][day].sum += int32(valueDelta)
			result[0][day].uniquePerson[s[0]] = true
		}
		return true
	})
	sort.Slice(sysKeys, func(i, j int) bool { return sysKeys[i] < sysKeys[j] })
	return result, sysKeys
}
