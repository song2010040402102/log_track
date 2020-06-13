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

var g_mapVideo map[int32]string = map[int32]string{
	0:   "所有",
	1:   "救济金双倍领取",
	2:   "没金币",
	3:   "金币场结束随机金币",
	4:   "看视频返还输的金币",
	5:   "看视频加成赢的金币",
	6:   "双倍红包券",
	7:   "下一局双倍红包券",
	8:   "下一局优先看底牌",
	9:   "看视频获得记牌器",
	10:  "金币场下局免扣金币",
	21:  "随机金币",
	22:  "开局看底牌",
	23:  "随机红包",
	24:  "goldSP(下一局)",
	25:  "goldSP(入口加号)",
	26:  "goldSP(入口icon)",
	27:  "goldSP(大厅加号)",
	28:  "goldSP(大厅金币不足)",
	29:  "goldSP(宝箱界面)",
	30:  "goldSP(来源toggle)",
	41:  "登录随机获得红包券",
	42:  "任务加倍奖励红包券",
	43:  "翻牌系统看视频得硬币",
	81:  "视频任务",
	82:  "视频任务并下载",
	98:  "纸巾机1",
	99:  "纸巾机2",
	100: "纸巾机3",
	101: "只赢不输",
	102: "签到看视频",
	103: "提现审核页看视频",
	201: "转盘",
	202: "大奖赛转盘",
	203: "大奖赛报名",
}

var g_videoStage []string = []string{"视频开始", "视频结束"}

type VideoInfo struct {
	sum      int32
	mapUsers map[string]bool
}

func VideoHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	videoType := getVideoType()
	mapVideo := getVideoInfo(gameId, startDate, endDate, channel, device)
	heads := make([]string, 0, len(g_videoStage)+1)
	heads = append(heads, "")
	for _, s := range g_videoStage {
		heads = append(heads, s)
	}
	childs := make([][]*TreeTable, 0, len(videoType))
	for _, vt := range videoType {
		childx := make([]*TreeTable, 0, len(g_videoStage)+1)
		childx = append(childx, NewTable(g_mapVideo[vt], nil, nil))
		for k, _ := range g_videoStage {
			heads := []string{"日期", "渠道", "设备", "总次数", "总人数", "人均次数"}
			items := make([][]string, 0, len(mapVideo[vt][k]))
			for i, v := range mapVideo[vt][k] {
				rows := make([]string, 0, len(heads))
				rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
				rows = append(rows, channel)
				rows = append(rows, device)
				rows = append(rows, strconv.Itoa(int(v.sum)))
				rows = append(rows, strconv.Itoa(len(v.mapUsers)))
				if len(v.mapUsers) > 0 {
					rows = append(rows, strconv.FormatFloat(float64(v.sum)/float64(len(v.mapUsers)), 'f', 2, 64))
				} else {
					rows = append(rows, "0.00")
				}
				items = append(items, rows)
			}
			childx = append(childx, NewTable("", heads, items))
		}
		childs = append(childs, childx)
	}
	return &ItemResult{
		TTable: NewTreeTable(fmt.Sprintf("看视频统计(%s ~ %s)", startDate, endDate), heads, childs),
	}
}

func getVideoInfo(gameId int32, startDate, endDate string, channel string, device string) map[int32][][]*VideoInfo {
	day := int(util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	mapVideo := make(map[int32][][]*VideoInfo)
	for t, _ := range g_mapVideo {
		vs := make([][]*VideoInfo, len(g_videoStage))
		for k := 0; k < len(g_videoStage); k++ {
			data := make([]*VideoInfo, day)
			for i := 0; i < day; i++ {
				data[i] = &VideoInfo{
					mapUsers: make(map[string]bool),
				}
			}
			vs[k] = data
		}
		mapVideo[t] = vs
	}
	handler := func(k int, day int32, s []string) bool {
		if len(s) < 11 {
			logs.Error("getVideoInfo data error!")
			return false
		}
		if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
			return true
		}
		vt, _ := strconv.ParseInt(s[10], 10, 32)
		if _, ok := g_mapVideo[int32(vt)]; ok {
			mapVideo[int32(vt)][k][day].sum++
			mapVideo[int32(vt)][k][day].mapUsers[s[0]] = true
			mapVideo[0][k][day].sum++
			mapVideo[0][k][day].mapUsers[s[0]] = true
		}
		return true
	}
	err := data.GetLogData(gameId, "VideoStartFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		return handler(0, day, s)
	})
	if err != nil {
		return nil
	}
	err = data.GetLogData(gameId, "VideoEndFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		return handler(1, day, s)
	})
	if err != nil {
		return nil
	}
	return mapVideo
}

func getVideoType() []int32 {
	vt := make([]int32, 0, len(g_mapVideo))
	for k, _ := range g_mapVideo {
		vt = append(vt, k)
	}
	sort.Slice(vt, func(i, j int) bool { return vt[i] < vt[j] })
	return vt
}
