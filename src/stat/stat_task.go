package stat

import (
	"common"
	"data"
	"fmt"
	"sort"
	"strconv"
	"util"

	"github.com/astaxie/beego/logs"
)

var taskState = []string{
	"触发任务",
	"完成任务",
	"领取任务",
}

var taskType = map[int32]string{
	1: "日常任务",
	2: "主线任务",
	3: "分享任务",
	4: "新手任务",
	5: "每周任务",
	6: "提现审核任务",
	// 7: "跳转任务"
}

func TaskHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	data := getTaskData(gameId, startDate, endDate, channel, device)
	days := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	dateStr := make([]string, days)
	for i := 0; i < days; i++ {
		dateStr[i] = util.Ts2date(util.Date2ts(startDate) + int64(i*86400))
	}
	taskHeads := append([]string{""}, taskState...)
	// tableHeads := []string{"日期", "渠道", "设备", "总次数"}

	childs := make([][]*TreeTable, 0, len(taskState))

	for j := 1; j <= len(taskType); j++ {
		var childsx []*TreeTable
		typeNum := int32(j)
		childsx = append(childsx, NewTable(taskType[typeNum], nil, nil))

		for i := 0; i < len(data); i++ {
			idChilds := make([][]*TreeTable, 1)
			var idkeys []int32
			for key := range data[i][typeNum] {
				idkeys = append(idkeys, key)
			}
			sort.Slice(idkeys, func(a, b int) bool { return idkeys[a] < idkeys[b] })
			var idStrs []string
			for _, id := range idkeys {
				idStrs = append(idStrs, strconv.Itoa(int(id)))
			}

			for _, k := range idkeys {
				var items [][]string
				taskIDTimes := data[i][typeNum][k]
				for d, times := range taskIDTimes {
					row := []string{
						dateStr[d],
						channel,
						device,
						strconv.Itoa(int(times)),
					}
					items = append(items, row)
				}
				tableHeads := []string{"日期", "渠道", "设备", "总次数"}
				idChilds[0] = append(idChilds[0], NewTable("", tableHeads, items))
			}
			childsx = append(childsx, NewTreeTable("", idStrs, idChilds))
		}
		childs = append(childs, childsx)
	}

	return &ItemResult{
		TTable: NewTreeTable(fmt.Sprintf("任务统计(%s ~ %s)", startDate, endDate), taskHeads, childs),
	}
}

func getTaskData(gameId int32, startDate, endDate, channel, device string) []map[int32]map[int32][]int32 {
	// return type: state-type-id-date-taskTimes :<
	result := make([]map[int32]map[int32][]int32, 3)
	for i := 0; i < 3; i++ {
		result[i] = make(map[int32]map[int32][]int32)
		for k := range taskType {
			result[i][k] = make(map[int32][]int32)
		}
	}
	days := (util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1

	handler := func(k int, day int32, s []string) bool {
		if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
			return true
		}
		if (k == 1 || k == 2) && len(s) < 12 || k == 0 && len(s) < 14 {
			logs.Error("invalid taskData length", k, len(s))
			return false
		}

		type1, _ := strconv.Atoi(s[10])
		t := int32(type1)
		if _, ok := taskType[t]; !ok {
			logs.Error("task type does not exist", t)
			return true
		}
		id, _ := strconv.Atoi(s[11])
		taskID := int32(id)
		if _, ok := result[k][t][taskID]; !ok {
			result[k][t][taskID] = make([]int32, days)
		}
		result[k][t][taskID][day]++
		return true
	}
	err := data.GetLogData(gameId, "TaskFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		return handler(0, day, s)
	})
	if err != nil {
		return nil
	}
	err = data.GetLogData(gameId, "TaskFinishFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		return handler(1, day, s)
	})
	if err != nil {
		return nil
	}
	err = data.GetLogData(gameId, "TaskDrawFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		return handler(2, day, s)
	})
	if err != nil {
		return nil
	}
	return result
}
