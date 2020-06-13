package stat

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
	"util"
)

func KeepAliveHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	if (util.Date2ts(endDate)-util.Date2ts(startDate))/86400 <= 0 {
		logs.Error("KeepAliveHandler, days must > 0!")
		return nil
	}
	var creates [3][]map[string]bool
	creates[0], _, _, _ = GetCreate(gameId, startDate, util.Ts2date(util.Date2ts(endDate)-86400), channel, device)
	daus := GetDAU(gameId, util.Ts2date(util.Date2ts(startDate)+86400), endDate, channel, device)
	plays := getPlay(gameId, startDate, util.Ts2date(util.Date2ts(endDate)-86400), channel, device, true)
	if len(creates[0]) != len(daus) || len(daus) != len(plays) {
		logs.Error("KeepAliveHandler, data length error!")
		return nil
	}
	for i := 1; i < len(creates); i++ {
		creates[i] = make([]map[string]bool, len(creates[0]))
		for j := 0; j < len(creates[i]); j++ {
			creates[i][j] = make(map[string]bool)
		}
	}
	for i := 0; i < len(creates[0]); i++ {
		for k, _ := range creates[0][i] {
			if _, ok := plays[i][k]; ok {
				creates[1][i][k] = true
			} else {
				creates[2][i][k] = true
			}
		}
	}
	heads := []string{"all", "打牌留存", "不打牌留存"}
	childs := make([][]*TreeTable, 1)
	childs[0] = make([]*TreeTable, 0, len(heads))
	for _, create := range creates {
		heads := []string{"日期", "渠道", "设备", "创建数"}
		for i := 0; i < len(create); i++ {
			heads = append(heads, fmt.Sprintf("%d留", i+1))
		}
		items := make([][]string, 0, len(create))
		for i, mc := range create {
			rows := make([]string, 0, len(heads))
			rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
			rows = append(rows, channel)
			rows = append(rows, device)
			rows = append(rows, strconv.Itoa(len(mc)))
			for j := 0; j < len(daus); j++ {
				if j+i < len(daus) {
					sum := 0
					for k, _ := range mc {
						if _, ok := daus[j+i][k]; ok {
							sum++
						}
					}
					if len(mc) > 0 {
						rows = append(rows, fmt.Sprintf("%.2f%%", float32(sum)*100/float32(len(mc))))
					} else {
						rows = append(rows, "0.00%")
					}
				} else {
					rows = append(rows, "0.00%")
				}
			}
			items = append(items, rows)
		}
		childs[0] = append(childs[0], NewTable("", heads, items))
	}
	return &ItemResult{
		TTable: NewTreeTable(fmt.Sprintf("用户留存统计(%s ~ %s)", startDate, endDate), heads, childs),
	}
}
