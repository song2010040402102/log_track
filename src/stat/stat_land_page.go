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

func LandPageHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	lpItems := getLandPage(gameId, startDate, endDate, channel, device)
	heads := []string{"日期", "渠道", "设备", "页面点击量", "加载完成量", "下载点击量", "页面点击量(去重)", "加载完成量(去重)", "下载点击量(去重)"}
	items := make([][]string, 0, len(lpItems))
	for i, v := range lpItems {
		rows := make([]string, 0, len(heads))
		rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
		rows = append(rows, channel)
		rows = append(rows, device)
		rows = append(rows, strconv.Itoa(int(v.InNum)))
		rows = append(rows, strconv.Itoa(int(v.OverNum)))
		rows = append(rows, strconv.Itoa(int(v.ClickNum)))
		rows = append(rows, strconv.Itoa(int(v.UInNum)))
		rows = append(rows, strconv.Itoa(int(v.UOverNum)))
		rows = append(rows, strconv.Itoa(int(v.UClickNum)))
		items = append(items, rows)
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("落地页统计(%s ~ %s)", startDate, endDate), heads, items),
	}
}

func getLandPage(gameId int32, startDate, endDate string, channel string, device string) []*common.LPItem {
	url := fmt.Sprintf("%s/land_page_res?channel=%s&device=%s&start=%d&end=%d",
		config.Get().Connect.Plat, url.QueryEscape(channel), device, util.Date2ts(startDate), util.Date2ts(endDate))
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("getLandPage http get error:", err, "url:", url)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("getLandPage ReadAll error:", err)
		return nil
	}
	lpRes := &common.LPRes{}
	err = util.FromJson(string(body), lpRes)
	if err != nil {
		logs.Error("getLandPage json parse error:", err)
		return nil
	}
	if lpRes.Ret != 0 {
		logs.Error("getLandPage result error with", lpRes.Ret)
		return nil
	}
	return lpRes.Items
}
