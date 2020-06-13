package stat

import (
	"common"
	"config"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"util"
)

func SellCountHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	res := getSellCount(gameId, startDate, endDate, channel, device)
	if res == nil {
		return nil
	}
	heads := []string{"", "进入盒子", "进入小游戏", "满足试玩时长"}
	childs := make([][]*TreeTable, 0, len(res.SellGames))
	for _, game := range res.SellGames {
		childx := make([]*TreeTable, 0, len(heads))
		childx = append(childx, NewTable(game.AppId, nil, nil))
		for k, sc := range game.SellCounts {
			heads := []string{"日期", "渠道", "设备", "次数", "人数", "人次比", "次数转化率", "人数转化率"}
			items := make([][]string, 0, len(sc))
			for i, v := range sc {
				rows := make([]string, 0, len(heads))
				if i == len(sc)-1 {
					rows = append(rows, "汇总")
				} else {
					rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
				}
				rows = append(rows, channel)
				rows = append(rows, device)
				rows = append(rows, strconv.Itoa(int(v.SumNum)))
				rows = append(rows, strconv.Itoa(int(v.PersonNum)))
				if v.SumNum > 0 {
					rows = append(rows, strconv.FormatFloat(float64(v.PersonNum)/float64(v.SumNum), 'f', 2, 64))
				} else {
					rows = append(rows, "0.00")
				}
				if k == 0 {
					rows = append(rows, "-")
					rows = append(rows, "-")
				} else {
					if game.SellCounts[k-1][i].SumNum > 0 {
						rows = append(rows, fmt.Sprintf("%.2f%%", float32(v.SumNum)*100/float32(game.SellCounts[k-1][i].SumNum)))
					} else {
						rows = append(rows, "0")
					}
					if game.SellCounts[k-1][i].PersonNum > 0 {
						rows = append(rows, fmt.Sprintf("%.2f%%", float32(v.PersonNum)*100/float32(game.SellCounts[k-1][i].PersonNum)))
					} else {
						rows = append(rows, "0")
					}
				}
				items = append(items, rows)
			}
			childx = append(childx, NewTable("", heads, items))
		}
		childs = append(childs, childx)
	}
	return &ItemResult{
		TTable: NewTreeTable(fmt.Sprintf("小游戏卖量统计(%s ~ %s)", startDate, endDate), heads, childs),
	}
}

func getSellCount(gameId int32, startDate, endDate string, channel string, device string) *common.SellGameRes {
	url := fmt.Sprintf("%s/sell_res?game_id=%d&item=sell_game&channel=%s&device=%s&start=%d&end=%d",
		config.Get().Connect.Plat, gameId, url.QueryEscape(channel), device, util.Date2ts(startDate), util.Date2ts(endDate))
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("getSellCount http get error:", err, "url:", url)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("getSellCount ReadAll error:", err)
		return nil
	}
	res := &common.SellGameRes{}
	err = util.FromJson(string(body), res)
	if err != nil {
		logs.Error("getSellCount json parse error:", err)
		return nil
	}
	if res.Ret != 0 {
		logs.Error("getSellCount result error with", res.Ret)
		return nil
	}
	return res
}
