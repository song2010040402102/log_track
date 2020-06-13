package stat

import (
	"common"
	"data"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"util"

	"github.com/astaxie/beego/logs"
)

type cashInfo struct {
	sum          float64
	personTime   int32
	uniquePerson map[string]bool
}

var g_cashTypes = map[float64]string{
	1000: "其他",
	-1:   "all",
	0.3:  "0.3元",
	1:    "1元",
	5:    "5元",
	10:   "10元",
	30:   "30元",
}

// CashHandler 返回每日用户的提取情况
func CashHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	var cashOutScale float64 = float64(GetCashScaleByGameId(gameId))

	heads := []string{"日期", "渠道", "设备", "登录名", "提现总额", "提现次数", "平均每次提现", "角色名"}

	// slice of cashOut maps
	var allCashOut []map[string][]float64
	// date - roleName map
	allRoleName := make(map[string]map[string]string)
	// loginName - channel
	chanSource := make(map[string]string)
	// loginName - device
	deviceSource := make(map[string]string)

	var cashOut map[string][]float64
	var roleName map[string]string

	err := data.GetLogData(gameId, "IncomeFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos == 0 {
			cashOut = make(map[string][]float64)
			roleName = make(map[string]string)
		} else if pos > 0 {
			if len(s) < 12 {
				logs.Error("CashHandler wrong data length", len(s))
				return false
			}
			if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
				return true
			}
			c, _ := strconv.ParseFloat(strings.TrimSpace(s[11]), 64)
			cashOut[s[0]] = append(cashOut[s[0]], c/cashOutScale)
			roleName[s[0]] = s[3]
			chanSource[s[0]] = s[8]
			deviceSource[s[0]] = s[9]
		} else {
			allCashOut = append(allCashOut, cashOut)
			allRoleName[util.Ts2date(util.Date2ts(startDate)+int64(day*86400))] = roleName
		}
		return true
	})
	if err != nil {
		return nil
	}

	var items [][]string
	for i, mp := range allCashOut {
		for ln, cash := range mp {
			var sum float64
			for _, v := range cash {
				sum += v
			}
			avg := float64(sum) / float64(len(cash))
			date := util.Ts2date(util.Date2ts(startDate) + int64(i*86400))
			row := []string{
				date,
				chanSource[ln],
				deviceSource[ln],
				ln,
				strconv.FormatFloat(sum, 'f', 2, 64),
				strconv.FormatInt(int64(len(cash)), 10),
				strconv.FormatFloat(avg, 'f', 2, 64),
				allRoleName[date][ln],
			}
			items = append(items, row)
		}
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("提现统计(%s ~ %s)", startDate, endDate), heads, items),
	}
}

func CashSummaryHandlerOld(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	heads := []string{"日期", "渠道", "设备", "提现总额", "提现总人次", "平均每次提现", "提现人数", "平均每人提现"}

	var summaries [][]string

	cashOutScale := float64(GetCashScaleByGameId(gameId))
	var allPersonTime int32
	var allCash float64
	allUniquePerson := make(map[string]bool)

	var personTime int32
	var cashOutSum float64
	var uniquePerson map[string]bool
	err := data.GetLogData(gameId, "IncomeFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos == 0 {
			personTime = 0
			cashOutSum = 0
			uniquePerson = make(map[string]bool)
		} else if pos > 0 {
			if len(s) < 12 {
				logs.Error("CashHandler wrong data length", len(s))
				return false
			}
			if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
				return true
			}
			cash, _ := strconv.ParseFloat(strings.TrimSpace(s[11]), 64)
			cashOutSum += cash / cashOutScale
			personTime++
			uniquePerson[s[0]] = true

			allCash += cash / cashOutScale
			allPersonTime++
			allUniquePerson[s[0]] = true
		} else {
			var avgCashOut, avgCashPerson float64
			if personTime > 0 {
				avgCashOut = cashOutSum / float64(personTime)
				avgCashPerson = cashOutSum / float64(len(uniquePerson))
			}
			row := []string{util.Ts2date(util.Date2ts(startDate) + int64(day*86400)), channel, device,
				strconv.FormatFloat(cashOutSum, 'f', 2, 64),
				strconv.FormatInt(int64(personTime), 10),
				strconv.FormatFloat(avgCashOut, 'f', 2, 64),
				strconv.FormatInt(int64(len(uniquePerson)), 10),
				strconv.FormatFloat(avgCashPerson, 'f', 2, 64),
			}
			summaries = append(summaries, row)
		}
		return true
	})
	if err != nil {
		return nil
	}

	var avgallCash, avgallCashPerson float64
	if allPersonTime > 0 {
		avgallCash = allCash / float64(allPersonTime)
		avgallCashPerson = allCash / float64(len(allUniquePerson))
	}
	summaryRow := []string{"汇总", channel, device,
		strconv.FormatFloat(allCash, 'f', 2, 64),
		strconv.FormatInt(int64(allPersonTime), 10),
		strconv.FormatFloat(avgallCash, 'f', 2, 64),
		strconv.FormatInt(int64(len(allUniquePerson)), 10),
		strconv.FormatFloat(avgallCashPerson, 'f', 2, 64),
	}
	summaries = append(summaries, summaryRow)

	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("提现汇总(%s ~ %s)", startDate, endDate), heads, summaries),
	}
}

// CashSummaryHandler 提供提现的汇总数据
func CashSummaryHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	cashOutScale := float64(GetCashScaleByGameId(gameId))
	days := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	dateStr := make([]string, days)
	for i := 0; i < days; i++ {
		dateStr[i] = util.Ts2date(util.Date2ts(startDate) + int64(i*86400))
	}

	var tables []*TreeTable
	result := make(map[string][]*cashInfo)
	result[g_cashTypes[-1]] = make([]*cashInfo, days)
	for i := 0; i < days; i++ {
		result[g_cashTypes[-1]][i] = &cashInfo{
			sum:          0,
			personTime:   0,
			uniquePerson: make(map[string]bool),
		}
	}

	err := data.GetLogData(gameId, "IncomeFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		if len(s) < 12 {
			logs.Error("CashHandler wrong data length", len(s))
			return false
		}
		if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
			return true
		}
		cash, _ := strconv.ParseFloat(strings.TrimSpace(s[11]), 64)
		cash = cash / cashOutScale
		if k, ok := g_cashTypes[cash]; ok {
			if _, ok2 := result[k]; !ok2 {
				result[k] = make([]*cashInfo, days)
				for j := 0; j < days; j++ {
					result[k][j] = &cashInfo{
						sum:          0,
						personTime:   0,
						uniquePerson: make(map[string]bool),
					}
				}
			}
			result[k][day].personTime++
			result[k][day].sum += cash
			result[k][day].uniquePerson[s[0]] = true
		} else {
			// 其他
			if _, ok2 := result[g_cashTypes[1000]]; !ok2 {
				result[g_cashTypes[1000]] = make([]*cashInfo, days)
				for j := 0; j < days; j++ {
					result[g_cashTypes[1000]][j] = &cashInfo{
						sum:          0,
						personTime:   0,
						uniquePerson: make(map[string]bool),
					}
				}
			}
			result[g_cashTypes[1000]][day].personTime++
			result[g_cashTypes[1000]][day].sum += cash
			result[g_cashTypes[1000]][day].uniquePerson[s[0]] = true
		}
		result[g_cashTypes[-1]][day].personTime++
		result[g_cashTypes[-1]][day].sum += cash
		result[g_cashTypes[-1]][day].uniquePerson[s[0]] = true
		return true
	})
	if err != nil {
		return nil
	}

	keySlice := []float64{}
	for k := range g_cashTypes {
		keySlice = append(keySlice, k)
	}
	sort.Slice(keySlice, func(i, j int) bool { return keySlice[i] < keySlice[j] })
	sourceHeads := []string{}
	for _, t := range keySlice {
		sourceHeads = append(sourceHeads, g_cashTypes[t])
		var items [][]string
		heads := []string{"日期", "渠道", "设备", "提现总次数", "提现总人数", "提现总金额"}
		if _, ok := result[g_cashTypes[t]]; ok {
			for i := 0; i < len(result[g_cashTypes[t]]); i++ {
				cashInfo := result[g_cashTypes[t]][i]
				row := []string{
					dateStr[i],
					channel,
					device,
					strconv.FormatInt(int64(cashInfo.personTime), 10),
					strconv.FormatInt(int64(len(cashInfo.uniquePerson)), 10),
					strconv.FormatFloat(cashInfo.sum, 'f', 2, 64),
				}
				items = append(items, row)
			}
			tables = append(tables, NewTable("", heads, items))
		} else {
			tables = append(tables, &TreeTable{})
		}
	}
	return &ItemResult{
		TTable: NewTreeTable(fmt.Sprintf("提现汇总(%s ~ %s)", startDate, endDate), sourceHeads, [][]*TreeTable{tables}),
	}
}
