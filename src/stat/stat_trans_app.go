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

func TransAppHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	transItems := getTransApp(gameId, startDate, endDate, channel, device)
	heads := []string{"日期", "渠道", "设备", "下载量", "激活量", "注册量"}
	items := make([][]string, 0, len(transItems))
	for i, v := range transItems {
		rows := make([]string, 0, len(heads))
		rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
		rows = append(rows, channel)
		rows = append(rows, device)
		rows = append(rows, strconv.Itoa(int(v.Click)))
		rows = append(rows, strconv.Itoa(int(v.Active)))
		rows = append(rows, strconv.Itoa(int(v.Register)))
		items = append(items, rows)
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("app转化统计(%s ~ %s)", startDate, endDate), heads, items),
	}
}

func getTransApp(gameId int32, startDate, endDate string, channel string, device string) []*common.TransItem {
	url := fmt.Sprintf("%s/trans_res?channel=%s&device=%s&start=%d&end=%d",
		config.Get().Connect.Plat, url.QueryEscape(channel), device, util.Date2ts(startDate), util.Date2ts(endDate)+86399)
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("getTransApp http get error:", err, "url:", url)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("getTransApp ReadAll error:", err)
		return nil
	}
	transRes := &common.TransRes{}
	err = util.FromJson(string(body), transRes)
	if err != nil {
		logs.Error("getTransApp json parse error:", err)
		return nil
	}
	if transRes.Ret != 0 {
		logs.Error("getTransApp result error with", transRes.Ret)
		return nil
	}
	return transRes.Items
}
