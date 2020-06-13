package plat

import (
	"common"
	"db"
	"fmt"
	"github.com/astaxie/beego/logs"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"util"
)

const (
	TRANS_EVENT_ACTIVE   int8 = 1
	TRANS_EVENT_REGISTER int8 = 2
)

type PlatInfo struct {
	platId  string
	cid     string
	channel string
}

var g_platInfo []*PlatInfo = []*PlatInfo{
	&PlatInfo{"qtt", "4773501", "515120"},
	&PlatInfo{"qtt", "4773516", "515121"},
	&PlatInfo{"qtt", "4299981", "515121"},
	&PlatInfo{"qtt", "4299955", "515121"},
	&PlatInfo{"sdhz", "1251816", "515108"},
	&PlatInfo{"sdhz", "1251818", "515109"},
}

var g_mapOS map[string]uint8 = map[string]uint8{
	"android": 0,
	"ios":     1,
}

var g_mapDevice map[string]string = map[string]string{
	"android": "android",
	"ios":     "iphone",
}

func Device2OS(s string) uint8 {
	for k, v := range g_mapDevice {
		if v == s {
			return g_mapOS[k]
		}
	}
	return 0
}

func GetPlatInfoByCId(cid string) *PlatInfo {
	for _, v := range g_platInfo {
		if v.cid == cid {
			return v
		}
	}
	return nil
}

func GetPlatInfoByChannel(channel string) *PlatInfo {
	for _, v := range g_platInfo {
		if v.channel == channel {
			return v
		}
	}
	return nil
}

func ParsePlatInfo(cid string, r *http.Request) (os int8, deviceId string, ts int64, callback_url string) {
	platInfo := GetPlatInfoByCId(cid)
	if platInfo == nil {
		logs.Error("ParsePlatInfo, invalid cid", cid)
		return
	}
	osTmp, _ := strconv.ParseInt(r.Form.Get("os"), 10, 32)
	os = int8(osTmp)
	if os == 0 {
		if platInfo.platId == "qtt" {
			deviceId = r.Form.Get("imeimd5") + "##" + r.Form.Get("androididmd5")
		} else {
			deviceId = r.Form.Get("imei") + "##" + r.Form.Get("androidid")
		}
	} else {
		deviceId = r.Form.Get("idfa")
	}
	ts, _ = strconv.ParseInt(r.Form.Get("timestamp"), 10, 32)
	callback_url = r.Form.Get("callback_url")
	return
}

func AddLandPageData(clientIP, channel, device, event string) {
	sql := fmt.Sprintf("insert into land_page (time, client_ip, channel, device, %s) value(?,?,?,?,1) on duplicate key update %s = %s + 1", event, event, event)
	_, err := db.GetMySql().Exec(sql, util.GetDate(), clientIP, channel, device)
	if err != nil {
		logs.Error("AddLandPageData, insert failed with", err)
		return
	}
}

func GetLandPageRes(channel string, device string, start, end int64) *common.LPRes {
	chanSql := getChannelSql(channel)
	devSql := getDeviceSql(device)
	sqlCond := fmt.Sprintf("time between '%s' and '%s'", util.Ts2date(start), util.Ts2date(end))
	if chanSql != "" {
		sqlCond += " and "
		sqlCond += chanSql
	}
	if devSql != "" {
		sqlCond += " and "
		sqlCond += devSql
	}
	sql := "select time, into_num, into_over_num, click_download from land_page"
	if sqlCond != "" {
		sql += fmt.Sprintf(" where %s", sqlCond)
	}
	logs.Info("GetLandPageRes query sql:", sql)
	rows, err := db.GetMySql().Query(sql)
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if err != nil {
		logs.Error("GetLandPageRes, query failed with %s", err)
		return &common.LPRes{1, nil}
	}
	lpRes := &common.LPRes{0, nil}
	days := (end-start)/86400 + 1
	for i := int64(0); i < days; i++ {
		lpRes.Items = append(lpRes.Items, &common.LPItem{})
	}
	var sTime string
	var inNum, overNum, clickNum int32
	for rows.Next() {
		err = rows.Scan(&sTime, &inNum, &overNum, &clickNum)
		if err != nil {
			logs.Error("GetLandPageRes, scan failed with %s", err)
			return &common.LPRes{2, nil}
		}
		index := int(util.Date2ts(sTime)-start) / 86400
		if index < 0 || index >= len(lpRes.Items) {
			logs.Error("GetLandPageRes, sTime abnormal!")
			continue
		}
		lpRes.Items[index].InNum += inNum
		lpRes.Items[index].OverNum += overNum
		lpRes.Items[index].ClickNum += clickNum
		if inNum > 0 {
			lpRes.Items[index].UInNum++
		}
		if overNum > 0 {
			lpRes.Items[index].UOverNum++
		}
		if clickNum > 0 {
			lpRes.Items[index].UClickNum++
		}
	}
	return lpRes
}

func insertTransData(channel string, os int8, deviceId string, ts int64, callback_url string, event int8) {
	//导量场景相比秒杀场景平缓得多，暂不加数据库缓冲机制
	_, err := db.GetMySql().Exec("insert into app_trans(channel, os, device_id, ts, callback_url, event) value(?,?,?,?,?,?)", channel, os, deviceId, ts, callback_url, event)
	if err != nil {
		logs.Error("insertTransData, insert failed with", err)
		return
	}
}

func AddTransAppData(cid string, os int8, deviceId string, ts int64, callback_url string) {
	platInfo := GetPlatInfoByCId(cid)
	if platInfo == nil {
		logs.Error("AddTransData, invalid cid", cid)
		return
	}
	callback_url, err := url.QueryUnescape(callback_url)
	if err != nil {
		logs.Error("AddTransData, url decode failed with", err)
		return
	}
	insertTransData(platInfo.channel, os, deviceId, ts, callback_url, 0)
}

func UpdateTransEvent(channel string, os int8, deviceId string, event int8) {
	if event != TRANS_EVENT_ACTIVE && event != TRANS_EVENT_REGISTER {
		logs.Error("UpdateTransEvent, event invalid!", event)
		return
	}
	var preEvent int8
	var callback_url string
	row := db.GetMySql().QueryRow("select callback_url, event from app_trans where channel=? and os=? and device_id=?", channel, os, deviceId)
	if err := row.Scan(&callback_url, &preEvent); err != nil {
		logs.Error("UpdateTransEvent, scan failed with", err)
		insertTransData(channel, os, deviceId, time.Now().Unix(), "", event)
		return
	}
	if preEvent >= event {
		logs.Error("UpdateTransEvent, event invalid!", preEvent, event)
		return
	}
	if _, err := db.GetMySql().Exec("update app_trans set event=? where channel=? and os=? and device_id=?", event, channel, os, deviceId); err != nil {
		//logs.Error("UpdateTransEvent, update failed with", err)
		return
	}
	if callback_url != "" {
		var para string
		if platInfo := GetPlatInfoByChannel(channel); platInfo != nil {
			if platInfo.platId == "qtt" {
				para = fmt.Sprintf("&op2=%d", event-1)
			} else if platInfo.platId == "sdhz" {
				if deviceId[:2] == "##" {
					para = "&dtype=androidid"
				} else {
					para = "&dtype=imei"
				}
				if event == TRANS_EVENT_ACTIVE {
					para += "&event_type=activate"
				} else {
					para += "&event_type=register"
				}
				para += fmt.Sprintf("&convert_time=%d", time.Now().Unix())
			}
		}
		if para != "" {
			resp, err := http.Get(callback_url + para)
			if err != nil {
				logs.Error("UpdateTransEvent http get error:", err, "url:", callback_url)
				return
			}
			if resp.StatusCode != 200 {
				logs.Error("UpdateTransEvent http status code:", resp.StatusCode)
				return
			}
			defer resp.Body.Close()
		}
	}
}

func GetTransRes(channel string, device string, start, end int64) *common.TransRes {
	chanSql := getChannelSql(channel)
	devSql := getDeviceSql2(device)
	sqlCond := fmt.Sprintf("ts between %d and  %d", start, end)
	if chanSql != "" {
		sqlCond += " and "
		sqlCond += chanSql
	}
	if devSql != "" {
		sqlCond += " and "
		sqlCond += devSql
	}
	sql := "select ts, event from app_trans"
	if sqlCond != "" {
		sql += fmt.Sprintf(" where %s", sqlCond)
	}
	logs.Info("GetTransRes query sql:", sql)
	rows, err := db.GetMySql().Query(sql)
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if err != nil {
		logs.Error("GetTransRes, query failed with %s", err)
		return &common.TransRes{1, nil}
	}

	type TsEvent struct {
		ts    int64
		event int8
	}
	var sEvent []*TsEvent
	for rows.Next() {
		te := &TsEvent{}
		err = rows.Scan(&te.ts, &te.event)
		if err != nil {
			logs.Error("GetTransRes, scan failed with %s", err)
			return &common.TransRes{2, nil}
		}
		sEvent = append(sEvent, te)
	}
	if len(sEvent) > 1 {
		sort.Slice(sEvent, func(i, j int) bool { return sEvent[i].ts <= sEvent[j].ts })
	}
	transRes := &common.TransRes{0, nil}
	days := (end-start)/86400 + 1
	for i := int64(0); i < days; i++ {
		transRes.Items = append(transRes.Items, &common.TransItem{})
	}
	for _, v := range sEvent {
		index := int(v.ts-start) / 86400
		if index < 0 || index >= len(transRes.Items) {
			logs.Error("GetTransRes, ts abnormal!")
			continue
		}
		if v.event >= TRANS_EVENT_REGISTER {
			transRes.Items[index].Register++
		}
		if v.event >= TRANS_EVENT_ACTIVE {
			transRes.Items[index].Active++
		}
		transRes.Items[index].Click++
	}
	return transRes
}

func getChannelSql(channel string) string {
	var cond string
	if channel != "" && channel != "all" {
		sChan := strings.Split(channel, ",")
		var oneChans []string
		var twoChans [][2]string
		for _, c := range sChan {
			sc := strings.Split(c, "~")
			if len(sc) == 2 {
				twoChans = append(twoChans, [2]string{sc[0], sc[1]})
			} else {
				oneChans = append(oneChans, c)
			}
		}
		cond = "("
		if len(oneChans) > 0 {
			cond += "channel in ('" + oneChans[0] + "'"
			for i := 1; i < len(oneChans); i++ {
				cond += ",'" + sChan[i] + "'"
			}
			cond += ")"
		}
		if len(twoChans) > 0 {
			if len(oneChans) == 0 {
				cond += "channel between '" + twoChans[0][0] + "' and '" + twoChans[0][1] + "'"
				for i := 1; i < len(twoChans); i++ {
					cond += " or channel between '" + twoChans[i][0] + "' and '" + twoChans[i][1] + "'"
				}
			} else {
				for _, v := range twoChans {
					cond += " or channel between '" + v[0] + "' and '" + v[1] + "'"
				}
			}
		}
		cond += ")"
	}
	return cond
}

func getDeviceSql(device string) string {
	var cond string
	if device != "" && device != "all" {
		sDev := strings.Split(device, ",")
		cond = "lower(substr(device, 1, instr(device, '_')-1)) in ("
		cond += "'" + g_mapDevice[sDev[0]] + "'"
		for i := 1; i < len(sDev); i++ {
			cond += ",'" + g_mapDevice[sDev[i]] + "'"
		}
		cond += ")"
	}
	return cond
}

func getDeviceSql2(device string) string {
	var cond string
	if device != "" && device != "all" {
		sDev := strings.Split(device, ",")
		cond = "os in ("
		cond += strconv.Itoa(int(g_mapOS[sDev[0]]))
		for i := 1; i < len(sDev); i++ {
			cond += "," + strconv.Itoa(int(g_mapOS[sDev[i]]))
		}
		cond += ")"
	}
	return cond
}
