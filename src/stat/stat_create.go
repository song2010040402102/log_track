package stat

import (
	"common"
	"data"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
	"time"
	"util"
)

func UserHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	now := time.Now().Unix()
	if end := util.Date2ts(endDate) + 86400; end < now {
		now = end
	}
	nums := make([]int32, (now-util.Date2ts(startDate))/3600+1)
	err := data.GetLogData(gameId, "Create", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		if len(s) < 10 {
			logs.Error("UserHandler data error!")
			return false
		}
		if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
			return true
		}
		ts, _ := strconv.ParseInt(s[6], 10, 32)
		index := int(ts-util.Date2ts(startDate)) / 3600
		if index >= 0 && index < len(nums)-1 {
			nums[index]++
		}
		nums[len(nums)-1]++
		return true
	})
	if err != nil {
		return nil
	}
	heads := []string{"时间", "渠道", "设备", "实际创建数"}
	items := make([][]string, 0, len(nums))
	for i, num := range nums {
		rows := make([]string, 0, len(heads))
		if i == len(nums)-1 {
			rows = append(rows, "汇总")
		} else {
			rows = append(rows, util.Ts2time(util.Date2ts(startDate)+int64(i*3600)))
		}
		rows = append(rows, channel)
		rows = append(rows, device)
		rows = append(rows, strconv.Itoa(int(num)))
		items = append(items, rows)
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("用户增长统计(%s ~ %s)", startDate, endDate), heads, items),
	}
}

func CreateHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	wxCreate := getWxCreateNum(gameId, startDate, endDate, channel, device)
	createAll, createChild, _, _ := GetCreate(gameId, startDate, endDate, channel, device)
	dau := GetDAU(gameId, startDate, endDate, channel, device)
	if len(wxCreate) != len(createAll) || len(createAll) != len(createChild) || len(createChild) != len(dau) {
		logs.Error("CreateHandler, data length error!")
		return nil
	}
	heads := []string{"日期", "渠道", "设备", "注册数", "实际创建数", "DAU", "裂变率", "裂变人数"}
	items := make([][]string, 0, len(wxCreate))
	for i, v := range wxCreate {
		rows := make([]string, 0, len(heads))
		rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
		rows = append(rows, channel)
		rows = append(rows, device)
		rows = append(rows, strconv.Itoa(int(v)))
		rows = append(rows, strconv.Itoa(len(createAll[i])))
		rows = append(rows, strconv.Itoa(len(dau[i])))
		if len(dau[i]) > 0 {
			rows = append(rows, fmt.Sprintf("%.2f%%", float32(len(createChild[i])*100)/float32(len(dau[i]))))
		} else {
			rows = append(rows, "0.00%")
		}
		rows = append(rows, strconv.Itoa(len(createChild[i])))
		items = append(items, rows)
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("创建、DAU、裂变统计(%s ~ %s)", startDate, endDate), heads, items),
	}
}

func getWxCreateNum(gameId int32, startDate, endDate string, channel string, device string) []int32 {
	ret := []int32{}
	var sum int32
	err := data.GetLogData(gameId, "WxCreate", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos == 0 {
			sum = 0
		} else if pos > 0 {
			if len(s) < 4 {
				logs.Error("getWxCreateNum data error!")
				return false
			}
			if !common.IsMultiCond(channel, s[2]) || !common.IsMultiCond(device, s[3]) {
				return true
			}
			sum++
		} else {
			ret = append(ret, sum)
		}
		return true
	})
	if err != nil {
		return nil
	}
	return ret
}
