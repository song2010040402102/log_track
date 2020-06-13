package stat

import (
	"common"
	"data"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
	"util"
)

type PayInfo struct {
	sumPay       int32
	sumPerson    int32
	sumNewPay    int32
	sumNewPerson int32
}

func PayHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	dau := GetDAU(gameId, startDate, endDate, channel, device)
	payInfo := getPayInfo(gameId, startDate, endDate, channel, device)
	if len(dau) != len(payInfo) {
		logs.Error("PayHandler, data length error!")
		return nil
	}
	heads := []string{"日期", "渠道", "设备", "DAU", "充值总额", "充值人数", "新增充值", "新增充值人数", "ARPU", "ARPPU"}
	items := make([][]string, 0, len(dau))
	for i, v := range payInfo {
		rows := make([]string, 0, len(heads))
		rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
		rows = append(rows, channel)
		rows = append(rows, device)
		rows = append(rows, strconv.Itoa(len(dau[i])))
		rows = append(rows, strconv.Itoa(int(v.sumPay)))
		rows = append(rows, strconv.Itoa(int(v.sumPerson)))
		rows = append(rows, strconv.Itoa(int(v.sumNewPay)))
		rows = append(rows, strconv.Itoa(int(v.sumNewPerson)))
		if len(dau[i]) > 0 {
			rows = append(rows, strconv.FormatFloat(float64(v.sumPay)/float64(len(dau[i])), 'f', 2, 64))
		} else {
			rows = append(rows, "0.00")
		}
		if v.sumPerson > 0 {
			rows = append(rows, strconv.FormatFloat(float64(v.sumPay)/float64(v.sumPerson), 'f', 2, 64))
		} else {
			rows = append(rows, "0.00")
		}
		items = append(items, rows)
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("付费统计(%s ~ %s)", startDate, endDate), heads, items),
	}
}

func getPayInfo(gameId int32, startDate, endDate string, channel string, device string) []*PayInfo {
	var payInfo *PayInfo
	ret := []*PayInfo{}
	data.GetLogData(gameId, "PayFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos == 0 {
			payInfo = &PayInfo{}
		} else if pos > 0 {
			if len(s) < 11 {
				logs.Error("getPayInfo data error!")
				return false
			}
			if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
				return true
			}
			cash, _ := strconv.ParseInt(s[10], 10, 32)
			payInfo.sumPay += int32(cash)
			payInfo.sumPerson += 1
			registerTime, _ := strconv.ParseInt(s[6], 10, 32)
			if util.Ts2date(util.Date2ts(startDate)+int64(day*86400)) == util.Ts2date(registerTime) {
				payInfo.sumNewPay += int32(cash)
				payInfo.sumNewPerson += 1
			}
		} else {
			ret = append(ret, payInfo)
		}
		return true
	})
	return ret
}
