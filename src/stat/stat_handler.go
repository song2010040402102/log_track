package stat

import (
	"clp"
	"common"
	"data"
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/astaxie/beego/logs"
	"sort"
	"strconv"
	"strings"
	"util"
)

type TreeTable struct {
	Title  string         `json:"title"`
	Heads  []string       `json:"heads"`
	Items  [][]string     `json:"items"`
	Childs [][]*TreeTable `json:"childs"`
}

func NewTable(title string, heads []string, items [][]string) *TreeTable {
	tt := &TreeTable{
		Title: title,
		Heads: heads,
		Items: items,
	}
	return tt
}

func NewTreeTable(title string, heads []string, childs [][]*TreeTable) *TreeTable {
	tt := &TreeTable{
		Title:  title,
		Heads:  heads,
		Childs: childs,
	}
	return tt
}

func IsEmptyTable(t *TreeTable) bool {
	return t == nil || t.Title == "" && len(t.Heads) == 0 && len(t.Items) == 0 && len(t.Childs) == 0
}

type ItemResult struct {
	Ret    int32      `json:"ret"`
	Msg    string     `json:"msg"`
	TTable *TreeTable `json:"ttable"`
}

func (ir *ItemResult) Error() string {
	if ir.Ret != 0 {
		return fmt.Sprintf("ret: %d, msg: %s", ir.Ret, ir.Msg)
	}
	return ""
}

var g_items []string = []string{
	"create", "pay", "all_play", "new_play", "keep_alive", "room_rule", "all_online_time", "new_online_time", "video", "gold_winlose",
	"gold_winlose_big", "place_winlose", "match", "grand_prix", "other_mini_game", "share_pic", "ddz_merge_play", "cash", "cash_summary",
	"redbag_output", "redbag_consume", "gold_video", "zjj_integal_output", "zjj_integal_consume", "diamond_output", "diamond_consume",
	"gold_output", "gold_consume", "enroll_voucher_output", "enroll_voucher_consume", "star_integal_output", "star_integal_consume",
	"wheel_output", "wheel_consume", "material_output", "material_consume", "task", "advertising", "land_page", "trans_app", "sell_count",
}

const SYS_ITEM_OFFSET = 14 * 26

var g_sysItems []string = []string{"sys_user"}

type pFunItemHandler func(int32, string, string, string, string) *ItemResult

var g_mapItemHandler map[string]pFunItemHandler = map[string]pFunItemHandler{
	"user":                   UserHandler,
	"create":                 CreateHandler,
	"pay":                    PayHandler,
	"all_play":               AllPlayHandler,
	"new_play":               NewPlayHandler,
	"keep_alive":             KeepAliveHandler,
	"room_rule":              RoomRuleHandler,
	"all_online_time":        AllOnlineTimeHandler,
	"new_online_time":        NewOnlineTimeHandler,
	"video":                  VideoHandler,
	"gold_winlose":           GoldWinLoseHandler,
	"gold_winlose_big":       GoldWinLoseBigHandler,
	"place_winlose":          PlaceWinLoseHandler,
	"match":                  MatchHandler,
	"grand_prix":             GrandPrixHandler,
	"other_mini_game":        OtherMiniGameHandler,
	"share_pic":              SharePicHandler,
	"ddz_merge_play":         DDZMergePlayHandler,
	"cash":                   CashHandler,
	"cash_summary":           CashSummaryHandler,
	"redbag_output":          RedBagOutPutHandler,
	"redbag_consume":         RedBagConsumeHandler,
	"gold_video":             GoldRoomVideoHandler,
	"zjj_integal_output":     ZJJIntegalOutputHandler,
	"zjj_integal_consume":    ZJJIntegalConsumeHandler,
	"diamond_output":         DiamondOutputHandler,
	"diamond_consume":        DiamondConsumeHandler,
	"gold_output":            GoldOutputHandler,
	"gold_consume":           GoldConsumeHandler,
	"enroll_voucher_output":  EnrollVoucherOutputHandler,
	"enroll_voucher_consume": EnrollVoucherConsumeHandler,
	"star_integal_output":    StarIntegalOutputHandler,
	"star_integal_consume":   StarIntegalConsumeHandler,
	"wheel_output":           WheelOutputHandler,
	"wheel_consume":          WheelConsumeHandler,
	"material_output":        MaterialOutputHandler,
	"material_consume":       MaterialConsumeHandler,
	"task":                   TaskHandler,
	"advertising":            AdvertisingHandler,
	"land_page":              LandPageHandler,
	"trans_app":              TransAppHandler,
	"sell_count":             SellCountHandler,

	//以下是为别的系统提供的接口
	"sys_user": SysUserHandler,
}

func GetItemResult(gameId int32, item string, startDate, endDate string, channel string, device string) *ItemResult {
	if pFun, _ := g_mapItemHandler[item]; pFun != nil {
		return pFun(gameId, startDate, endDate, util.RemoveAllBlank(channel), device)
	} else {
		logs.Error("GetItemResult", item, "invalid item!")
	}
	return nil
}

func GetItemResult2(gameId int32, items string, startDate, endDate string, channel string, device string) *ItemResult {
	items = util.RemoveBlank(items)
	channel = util.RemoveAllBlank(channel)
	res, err := getItemsRes(gameId, items, startDate, endDate, channel, device)
	if err != nil {
		return &ItemResult{
			Ret: 1,
			Msg: err.Error(),
		}
	}
	heads := []string{"日期", "渠道", "设备"}
	for _, v := range res {
		if len(v) > 0 {
			heads = append(heads, v[0])
		}
	}
	days := (util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	tItems := make([][]string, 0, days)
	for i := int64(0); i < days; i++ {
		rows := make([]string, 0, len(heads))
		rows = append(rows, util.Ts2date(util.Date2ts(startDate)+i*86400))
		rows = append(rows, channel)
		rows = append(rows, device)
		for _, v := range res {
			if int(i) < len(v)-1 {
				rows = append(rows, v[i+1])
			} else {
				rows = append(rows, "")
			}
		}
		tItems = append(tItems, rows)
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("%s(%s ~ %s)", items, startDate, endDate), heads, tItems),
	}
}

func WriteXLSX(res *ItemResult) (string, error) {
	if res.Ret != 0 {
		return "", errors.New(fmt.Sprintf("error code: %d, detail: %s", res.Ret, res.Msg))
	} else if res.TTable == nil {
		return "", errors.New("excel data empty!")
	}

	filename := fmt.Sprintf("temp_%d.xlsx", util.GetUUID())
	f := excelize.NewFile()

	mSize := make(map[*TreeTable]Size)
	getTableSize(res.TTable, mSize)
	writeTabel(f, res.TTable, mSize, 1, 0)

	if err := f.SaveAs("./download/" + filename); err != nil {
		logs.Error("SaveAs error:", err)
		return "", err
	}
	return filename, nil
}

type Size struct {
	X, Y   uint32
	XX, YY []uint32
}

func getTableSize(tt *TreeTable, mSize map[*TreeTable]Size) {
	if len(tt.Heads) == 0 && len(tt.Items) == 0 && len(tt.Childs) == 0 {
		return
	}
	if _, ok := mSize[tt]; ok {
		return
	}
	var x, y uint32
	var xx, yy []uint32
	if tt.Title != "" {
		x = 1
		y++
	}
	if len(tt.Heads) > 0 {
		x = uint32(len(tt.Heads))
		y++
	}
	if len(tt.Items) > 0 {
		x = uint32(len(tt.Heads))
		y += uint32(len(tt.Items))
	} else if len(tt.Childs) > 0 {
		xx = make([]uint32, len(tt.Heads))
		yy = make([]uint32, len(tt.Childs))
		for i, rows := range tt.Childs {
			for j, col := range rows {
				getTableSize(col, mSize)
				if s, ok := mSize[col]; ok {
					if xx[j] < s.X {
						xx[j] = s.X
					}
					if yy[i] < s.Y {
						yy[i] = s.Y
					}
				}
			}
		}
		x = 0
		for i := 0; i < len(xx); i++ {
			if xx[i] == 0 {
				xx[i] = 1
			}
			x += xx[i]
		}
		for i := 0; i < len(yy); i++ {
			if yy[i] == 0 {
				yy[i] = 1
			}
			y += yy[i]
		}
	}
	mSize[tt] = Size{X: x, Y: y, XX: xx, YY: yy}
}

func writeTabel(f *excelize.File, tt *TreeTable, mSize map[*TreeTable]Size, startX, startY uint32) {
	if tt == nil {
		return
	}
	size, ok := mSize[tt]
	if !ok {
		return
	}
	border := `"border":[{"type":"left","color":"000000","style":1},{"type":"top","color":"000000","style":1},{"type":"bottom","color":"000000","style":1},{"type":"right","color":"000000","style":1}]`
	fontT := `"font":{"bold":true,"family":"宋体","size":24}`
	fontH := `"font":{"bold":true,"family":"宋体","size":11}`
	fontC := `"font":{"family":"宋体","size":11}`
	align := `"alignment":{"horizontal":"center","vertical":"center"}`
	if tt.Title != "" {
		f.MergeCell("Sheet1", xy2str(startX, startY), xy2str(startX+size.X-1, startY))
		f.SetCellValue("Sheet1", xy2str(startX, startY), tt.Title)
		style, _ := f.NewStyle(fmt.Sprintf("{%s,%s}", fontT, align))
		f.SetCellStyle("Sheet1", xy2str(startX, startY), xy2str(startX+size.X-1, startY), style)
		f.SetRowHeight("Sheet1", int(startY+1), 31.5)
		startY++
	}
	if len(tt.Heads) > 0 {
		lastX := startX
		for i, v := range tt.Heads {
			w := uint32(1)
			if i < len(size.XX) && size.XX[i] > 1 {
				w = size.XX[i]
			}
			f.MergeCell("Sheet1", xy2str(lastX, startY), xy2str(lastX+w-1, startY))
			f.SetCellValue("Sheet1", xy2str(lastX, startY), v)
			lastX += w
		}
		style, _ := f.NewStyle(fmt.Sprintf("{%s,%s,%s}", border, fontH, align))
		f.SetCellStyle("Sheet1", xy2str(startX, startY), xy2str(startX+size.X-1, startY), style)
		startY++
	}
	if len(tt.Items) > 0 {
		for i, rows := range tt.Items {
			for j, col := range rows {
				if v1, err1 := strconv.ParseInt(col, 10, 64); err1 == nil {
					f.SetCellValue("Sheet1", xy2str(startX+uint32(j), startY+uint32(i)), v1)
				} else if v2, err2 := strconv.ParseFloat(col, 64); err2 == nil {
					f.SetCellValue("Sheet1", xy2str(startX+uint32(j), startY+uint32(i)), v2)
				} else {
					f.SetCellValue("Sheet1", xy2str(startX+uint32(j), startY+uint32(i)), col)
				}
			}
		}
		style, _ := f.NewStyle(fmt.Sprintf("{%s,%s,%s}", border, fontC, align))
		f.SetCellStyle("Sheet1", xy2str(startX, startY), xy2str(startX+size.X-1, startY+uint32(len(tt.Items)-1)), style)
	} else if len(tt.Childs) > 0 {
		lastY := startY
		for i, rows := range tt.Childs {
			lastX := startX
			h := uint32(1)
			if i < len(size.YY) && size.YY[i] > 1 {
				h = size.YY[i]
			}
			for j, col := range rows {
				w := uint32(1)
				if j < len(size.XX) && size.XX[j] > 1 {
					w = size.XX[j]
				}
				if col == nil {
					f.MergeCell("Sheet1", xy2str(lastX, lastY), xy2str(lastX+w-1, lastY+h-1))
				} else if len(col.Heads) == 0 && len(col.Items) == 0 && len(col.Childs) == 0 {
					f.MergeCell("Sheet1", xy2str(lastX, lastY), xy2str(lastX+w-1, lastY+h-1))
					style, _ := f.NewStyle(fmt.Sprintf("{%s,%s,%s}", border, fontH, align))
					f.SetCellStyle("Sheet1", xy2str(lastX, lastY), xy2str(lastX+w-1, lastY+h-1), style)
					f.SetCellValue("Sheet1", xy2str(lastX, lastY), col.Title)
				} else {
					writeTabel(f, col, mSize, lastX, lastY)
				}
				lastX += w
			}
			lastY += h
		}
	}
}

func xy2str(x, y uint32) string {
	return util.Dec2letter(x) + strconv.Itoa(int(y+1))
}

func getItemsRes(gameId int32, items string, startDate, endDate string, channel string, device string) (ret [][]string, err error) {
	data := make(map[string]*ItemResult)
	sItem := strings.Split(items, ";")
	for _, item := range sItem {
		if len(item) == 0 {
			continue
		}
		title := ""
		if item[0] == '"' {
			for i := 1; i < len(item); i++ {
				if item[i] == '"' {
					title = item[1:i]
					item = item[i+1:]
					break
				}
			}
			if title == "" {
				return ret, errors.New(item + " syntax error!")
			}
		}
		convert := 0
		if len(item) > 0 && item[0] == '%' {
			convert = 1
			item = item[1:]
		}
		exp, err := clp.ParseExpress(item)
		if err != nil {
			return ret, err
		}
		if exp.GetOpera() == clp.OP_NONE {
			res, err := getResByItem(data, gameId, item, startDate, endDate, channel, device)
			if err != nil {
				return ret, err
			}
			if title != "" && len(res) == 1 && len(res[0]) > 0 {
				res[0][0] = title
			}
			ret = append(ret, res...)
		} else {
			handler := func(s string) (vals []float64, err error) {
				days := (util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
				vals = make([]float64, days)
				val, err := strconv.ParseFloat(s, 32)
				if err == nil {
					for i := 0; i < len(vals); i++ {
						vals[i] = val
					}
				} else {
					res, err := getResByItem(data, gameId, s, startDate, endDate, channel, device)
					if err != nil {
						return vals, err
					}
					if len(res) > 1 {
						return vals, errors.New(s + " multi-values forbid arithmetic!")
					} else if len(res) == 0 {
						return vals, errors.New(s + " unknown error!")
					}
					for i := 1; i < len(res[0]); i++ {
						if i > len(vals) {
							continue
						}
						if res[0][i] == "" {
							vals[i-1] = 0
						} else {
							val, err := strconv.ParseInt(res[0][i], 10, 32)
							if err == nil {
								vals[i-1] = float64(val)
							} else {
								vals[i-1], err = strconv.ParseFloat(res[0][i], 32)
								if err != nil {
									return vals, err
								}
							}
						}
					}
				}
				return vals, nil
			}
			err = clp.RunExpress(exp, handler)
			if err != nil {
				return ret, err
			} else {
				res := make([]string, len(exp.GetValues())+1)
				res[0] = title
				for i, v := range exp.GetValues() {
					if convert == 1 {
						res[i+1] = fmt.Sprintf("%.2f%%", v*100)
					} else {
						res[i+1] = strconv.FormatFloat(v, 'f', 2, 64)
					}
				}
				ret = append(ret, res)
			}
		}
	}
	return ret, err
}

func getResByItem(data map[string]*ItemResult, gameId int32, item string, startDate, endDate string, channel string, device string) (ret [][]string, err error) {
	sItem := strings.Split(item, ".")
	if len(sItem) == 0 {
		return ret, errors.New("Unknown error!")
	}
	index := int(util.Letter2dec(sItem[0]))
	item = ""
	if index < SYS_ITEM_OFFSET {
		if index >= 0 && index < len(g_items) {
			item = g_items[index]
		}
	} else {
		index -= SYS_ITEM_OFFSET
		if index >= 0 && index < len(g_sysItems) {
			item = g_sysItems[index]
		}
	}
	if item == "" {
		return ret, errors.New(sItem[0] + "invalid!")
	}
	itemRes, _ := data[item]
	if itemRes == nil {
		itemRes = GetItemResult(gameId, item, startDate, endDate, channel, device)
		if itemRes == nil || IsEmptyTable(itemRes.TTable) {
			return ret, errors.New(item + " empty!")
		}
		data[item] = itemRes
	}
	return traverItemRes(itemRes.TTable, sItem, 1, "", ret)
}

func traverItemRes(tt *TreeTable, items []string, cur int, title string, data [][]string) (ret [][]string, err error) {
	s, e := [2]int{0, 0}, [2]int{-1, -1}
	for i := 0; i < 2; i++ {
		if cur+i >= len(items) {
			break
		}
		ss := strings.Split(items[cur+i], "~")
		if len(ss) <= 0 || len(ss) > 2 {
			return ret, errors.New(items[cur+i] + " syntax error!")
		}
		val := int64(0)
		if ss[0] != "" {
			val, err = strconv.ParseInt(ss[0], 10, 32)
			if err != nil {
				return ret, err
			}
		}
		if len(ss) == 1 {
			s[i], e[i] = int(val), int(val)
		} else {
			s[i] = int(val)
			if ss[1] != "" {
				val, err = strconv.ParseInt(ss[1], 10, 32)
				if err != nil {
					return ret, err
				}
				e[i] = int(val)
			}
		}
		if s[i] < 0 {
			return ret, errors.New("Start index cannot be negative!")
		}
	}
	if len(tt.Items) > 0 {
		s[0] += 3
		if e[0] == -1 {
			e[0] = len(tt.Heads) - 1
		} else {
			e[0] += 3
		}
		if e[0] >= len(tt.Heads) {
			e[0] = len(tt.Heads) - 1
		}
		for i := s[0]; i <= e[0]; i++ {
			col := make([]string, len(tt.Items)+1)
			col[0] = title + tt.Heads[i]
			for j := 0; j < len(tt.Items); j++ {
				col[j+1] = tt.Items[j][i]
			}
			data = append(data, col)
		}
	} else {
		if len(tt.Heads) > 0 {
			if tt.Heads[0] == "" {
				s[1]++
				if e[1] == -1 {
					e[1] = len(tt.Heads) - 1
				} else {
					e[1]++
				}
			}
		} else {
			return data, errors.New("Table no heads!")
		}
		if e[0] == -1 || e[0] >= len(tt.Childs) {
			e[0] = len(tt.Childs) - 1
		}
		if e[1] == -1 || e[1] >= len(tt.Heads) {
			e[1] = len(tt.Heads) - 1
		}
		for i := s[0]; i <= e[0]; i++ {
			rtitle := title
			if s[0] != e[0] && len(tt.Childs[i]) > 0 && tt.Childs[i][0] != nil {
				rtitle += tt.Childs[i][0].Title + "-"
			}
			for j := s[1]; j <= e[1]; j++ {
				if IsEmptyTable(tt.Childs[i][j]) {
					continue
				}
				ctitle := rtitle
				if s[1] != e[1] && j < len(tt.Heads) {
					ctitle += tt.Heads[j] + "-"
				}
				data, err = traverItemRes(tt.Childs[i][j], items, cur+2, ctitle, data)
				if err != nil {
					return data, err
				}
			}
		}
	}
	return data, err
}

func AddIndexToItemRes(itemRes *ItemResult) {
	if itemRes == nil {
		return
	}
	addIndexToIR(itemRes.TTable)
}

func addIndexToIR(tt *TreeTable) {
	if tt == nil {
		return
	}
	count := 0
	for i := 0; i < len(tt.Heads); i++ {
		if tt.Heads[i] == "" {
			continue
		}
		if len(tt.Items) > 0 && i < 3 {
			continue
		}
		tt.Heads[i] = fmt.Sprintf("%s(%d)", tt.Heads[i], count)
		count++
	}
	count = 0
	for i := 0; i < len(tt.Childs); i++ {
		if len(tt.Childs[i]) > 0 && tt.Childs[i][0].Title != "" && len(tt.Childs[i][0].Items) == 0 && len(tt.Childs[i][0].Childs) == 0 {
			tt.Childs[i][0].Title = fmt.Sprintf("%s(%d)", tt.Childs[i][0].Title, count)
			count++
		}
		for j := 0; j < len(tt.Childs[i]); j++ {
			addIndexToIR(tt.Childs[i][j])
		}
	}
}

func ParseLoginName(loginName string) (string, int32) {
	pid, ts, randId, inc, index, channel := int32(0), int32(0), int32(0), int32(0), int32(0), ""
	fmt.Sscanf(loginName, "%04x-%08x-%04x-%04x-%01x-%s", &pid, &ts, &randId, &inc, &index, &channel)
	return channel, index
}

func GetDAU(gameId int32, startDate, endDate string, channel string, device string) []map[string]bool {
	var mapUsers map[string]bool
	ret := []map[string]bool{}
	data.GetLogData(gameId, "Login", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos == 0 {
			mapUsers = make(map[string]bool)
		} else if pos > 0 {
			if len(s) < 10 {
				logs.Error("GetDAU data error!")
				return false
			}
			if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
				return true
			}
			mapUsers[s[0]] = true
		} else {
			ret = append(ret, mapUsers)
		}
		return true
	})
	return ret
}

func GetCreate(gameId int32, startDate, endDate string, channel string, device string) ([]map[string]bool, []map[string]bool, []map[string]bool, []map[string]bool) {
	var all, child, noShareTour, noShareAuth map[string]bool
	retAll, retChild, retNoShareTour, retNoShareAuth := []map[string]bool{}, []map[string]bool{}, []map[string]bool{}, []map[string]bool{}
	data.GetLogData(gameId, "Create", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos == 0 {
			all, child, noShareTour, noShareAuth = make(map[string]bool), make(map[string]bool), make(map[string]bool), make(map[string]bool)
		} else if pos > 0 {
			if len(s) < 10 {
				logs.Error("GetCreateNum data error!")
				return false
			}
			if !common.IsMultiCond(device, s[9]) {
				return true
			}
			if common.IsMultiCond(channel, s[8]) {
				all[s[0]] = true
			}
			if chnl, index := ParseLoginName(s[0]); index > 0 {
				if common.IsMultiCond(channel, chnl) {
					child[s[0]] = true
				}
			} else {
				if common.IsMultiCond(channel, s[8]) {
					if s[5] == "" {
						noShareTour[s[0]] = true
					} else {
						noShareAuth[s[0]] = true
					}
				}
			}
		} else {
			retAll = append(retAll, all)
			retChild = append(retChild, child)
			retNoShareTour = append(retNoShareTour, noShareTour)
			retNoShareAuth = append(retNoShareAuth, noShareAuth)
		}
		return true
	})
	return retAll, retChild, retNoShareTour, retNoShareAuth
}

func GameHasRoom(gameId int32, roomId int32) bool {
	if gameId == common.MJ_51 || gameId == common.MJ_YL || gameId == common.MJ_YL_ZJ {
		return roomId >= ROOM_ALL && roomId <= 3003
	} else if gameId == common.DDZ_LY || gameId == common.DDZ_LQ || gameId == common.DDZ_XM || gameId == common.DDZ_CS {
		return roomId == ROOM_GOLD
	} else if gameId == common.DDZ_WX_CS || gameId == common.DDZ_WX {
		return roomId == ROOM_ALL || roomId == ROOM_NORMAL || roomId == ROOM_GOLD || roomId == ROOM_PLACE ||
			roomId >= 5001 && roomId <= 6003
	}
	return false
}

func GameHasRule(gameId int32, ruleId int32) bool {
	if gameId == common.MJ_51 || gameId == common.MJ_YL || gameId == common.MJ_YL_ZJ {
		return ruleId == 0 || ruleId >= 4001 && ruleId < 40000
	} else if gameId == common.DDZ_LY || gameId == common.DDZ_LQ || gameId == common.DDZ_XM ||
		gameId == common.DDZ_CS || gameId == common.DDZ_WX_CS || gameId == common.DDZ_WX {
		return ruleId == 0 || ruleId == 1000 || ruleId == 1001 || ruleId == 1002
	}
	return false
}

func GetCashScaleByGameId(gameId int32) int32 {
	if gameId == common.MJ_51 || gameId == common.MJ_YL || gameId == common.MJ_YL_ZJ || gameId == common.DDZ_WX_CS || gameId == common.DDZ_WX {
		return 100
	} else if gameId == common.DDZ_LY || gameId == common.DDZ_LQ || gameId == common.DDZ_XM || gameId == common.DDZ_CS {
		return 10000
	}
	return 1
}

func GetRoomIds(gameId int32) []int32 {
	roomIds := []int32{}
	for k, _ := range g_mapRoom {
		if GameHasRoom(gameId, k) {
			roomIds = append(roomIds, k)
		}
	}
	sort.Slice(roomIds, func(i, j int) bool { return roomIds[i] < roomIds[j] })
	return roomIds
}

func GetRuleIds(gameId int32) []int32 {
	ruleIds := []int32{}
	for k, _ := range g_mapRule {
		if GameHasRule(gameId, k) {
			ruleIds = append(ruleIds, k)
		}
	}
	sort.Slice(ruleIds, func(i, j int) bool { return ruleIds[i] < ruleIds[j] })
	return ruleIds
}
