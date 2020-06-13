package plat

import (
	"bridge"
	"cache"
	"common"
	"db"
	"fmt"
	"github.com/astaxie/beego/logs"
	"sync"
	"time"
	"util"
)

func InitSellEvent() {
	g_mapEnterGame = make(map[string]int64)
	util.GobDecodeFile("tmp/sell_event.tmp", &g_mapEnterGame)

	g_mapAward = make(map[string]bool)
	util.GobDecodeFile("tmp/sell_award.tmp", &g_mapAward)

	cache.GetCacheManager().Set(common.CACHE_TYPE_MYSQL, 0, 0)
	cache.GetCacheManager().Set(common.CACHE_TYPE_HTTP, 0, 0)

	//AutoClearAward()
}

func SaveSellEvent() {
	util.GobEncodeFile("tmp/sell_event.tmp", g_mapEnterGame)
	util.GobEncodeFile("tmp/sell_award.tmp", g_mapAward)
}

func AddSellEvent(loginname string, channel string, os uint8, miniAppId, miniGameId string, event uint16) {
	now := time.Now().Unix()
	gameId := common.GetGameIdByChannel(channel)
	if event == common.SELL_EVENT_GAME_ENTER {
		g_eventLock.Lock()
		g_mapEnterGame[loginname] = now
		g_eventLock.Unlock()
	} else if event == common.SELL_EVENT_GAME_EXIT {
		g_eventLock.Lock()
		start := g_mapEnterGame[loginname]
		delete(g_mapEnterGame, loginname)
		g_eventLock.Unlock()
		if start == 0 {
			logs.Warning("AddSellEvent, not enter game!")
		}
		if cfg := GetSellCfgById(miniAppId, miniGameId); cfg != nil {
			g_awardLock.Lock()
			_, ok := g_mapAward[loginname+miniAppId+miniGameId]
			g_awardLock.Unlock()
			if !ok {
				if uint16(now-start) >= cfg.StayTime {
					cache.GetCacheManager().AddCache2(common.CACHE_TYPE_HTTP, &SellEventAward{
						gameId:     gameId,
						miniAppId:  miniAppId,
						miniGameId: miniGameId,
						loginname:  loginname,
						award:      cfg.Award,
					})
				} else {
					logs.Notice("AddSellEvent", now-start, "time not enough!")
				}
			} else {
				logs.Notice("AddSellEvent", "award has got!")
			}
		} else {
			logs.Error("AddSellEvent", miniAppId, miniGameId, "not cfg!")
		}
	}
	cache.GetCacheManager().AddCache2(common.CACHE_TYPE_MYSQL, &SellEventInsert{
		gameId:     gameId,
		loginname:  loginname,
		channel:    channel,
		os:         os,
		miniAppId:  miniAppId,
		miniGameId: miniGameId,
		event:      event,
		ts:         now,
	})
}

type SellEventInsert struct {
	gameId     int32
	loginname  string
	channel    string
	os         uint8
	miniAppId  string
	miniGameId string
	event      uint16
	ts         int64
}

func (si *SellEventInsert) Run() {
	_, err := db.GetMySql().Exec("insert into sell_mini_game_event (game_id, loginname, channel, os, mini_app_id, mini_game_id, event, ts) value(?,?,?,?,?,?,?,?)",
		si.gameId, si.loginname, si.channel, si.os, si.miniAppId, si.miniGameId, si.event, si.ts)
	if err != nil {
		logs.Error("SellEventInsert, insert failed with", err)
	}
}

type SellEventAward struct {
	gameId     int32
	miniAppId  string
	miniGameId string
	loginname  string
	award      uint32
}

func (sa *SellEventAward) Run() {
	mPara := make(map[string]string)
	mPara["loginname"] = sa.loginname
	mPara["do"] = "108"
	mPara["what"] = sa.miniAppId + "##" + sa.miniGameId
	mPara["award"] = fmt.Sprintf("4,4,%d", sa.award)
	resp := bridge.HttpPostServer2(sa.gameId, "award_inform", mPara)
	g_awardLock.Lock()
	g_mapAward[sa.loginname+sa.miniAppId+sa.miniGameId] = true
	g_awardLock.Unlock()
	logs.Info("SellEventAward, loginname:", sa.loginname, "award:", sa.award, "resp:", resp)
}

func GetAwardIds(loginname string) []string {
	var ret []string
	cfgs := GetSellCfg()
	g_awardLock.Lock()
	for _, v := range cfgs {
		if v != nil && g_mapAward[loginname+v.MiniAppId+v.MiniGameId] {
			ret = append(ret, v.MiniAppId+v.MiniGameId)
		}
	}
	g_awardLock.Unlock()
	return ret
}

func AutoClearAward() {
	const days = 1
	t := time.Now()
	n := t.Hour()*3600 + t.Minute()*60 + t.Second()
	timer := time.NewTimer(time.Duration(days*24*3600-n) * time.Second)
	go func(t *time.Timer) {
		for {
			<-t.C
			g_awardLock.Lock()
			g_mapAward = make(map[string]bool)
			g_awardLock.Unlock()
			t.Reset(days * 24 * 3600 * time.Second)
		}
	}(timer)
}

type SellData struct {
	loginname  string
	miniAppId  string
	miniGameId string
	event      uint16
	ts         int64
}

func getSellData(gameId int32, channel, device string, start, end int64) []*SellData {
	var ret []*SellData
	chanSql := getChannelSql(channel)
	devSql := getDeviceSql2(device)
	sqlCond := fmt.Sprintf("game_id=%d and ts between %d and  %d", gameId, start, end)
	if chanSql != "" {
		sqlCond += " and "
		sqlCond += chanSql
	}
	if devSql != "" {
		sqlCond += " and "
		sqlCond += devSql
	}
	sql := "select loginname, mini_app_id, mini_game_id, event, ts from sell_mini_game_event"
	if sqlCond != "" {
		sql += fmt.Sprintf(" where %s", sqlCond)
	}
	logs.Info("getSellGame query sql:", sql)
	rows, err := db.GetMySql().Query(sql)
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if err != nil {
		logs.Error("getSellGame, query failed with %s", err)
		return ret
	}

	mLastGo := make(map[string]int64)
	for rows.Next() {
		sd := &SellData{}
		err = rows.Scan(&sd.loginname, &sd.miniAppId, &sd.miniGameId, &sd.event, &sd.ts)
		if err != nil {
			logs.Error("getSellGame, scan failed with %s", err)
			return ret
		}
		cfg := GetSellCfgById(sd.miniAppId, sd.miniGameId)
		if cfg == nil {
			continue
		}
		if sd.event == common.SELL_EVENT_GAME_ENTER {
			mLastGo[sd.loginname+sd.miniAppId+sd.miniGameId] = sd.ts
		} else if sd.event == common.SELL_EVENT_GAME_EXIT {
			if sd.ts-mLastGo[sd.loginname+sd.miniAppId+sd.miniGameId] < int64(cfg.StayTime) {
				continue
			}
		}
		ret = append(ret, sd)
	}
	return ret
}

func GetSellRes(gameId int32, item, channel, device string, start, end int64) interface{} {
	switch item {
	case "sell_game":
		return getSellGame(gameId, channel, device, start, end)
	case "sell_user":
		return getSellUser(gameId, channel, device, start, end)
	}
	return nil
}

func getSellGame(gameId int32, channel, device string, start, end int64) *common.SellGameRes {
	sgr := &common.SellGameRes{Ret: 0}
	days := int(end-start)/86400 + 1
	allIds := []string{"all"}
	allIds = append(allIds, GetAllMiniAppIds()...)
	allIds = append(allIds, GetAllMiniGameIds()...)
	for _, v := range allIds {
		game := &common.SellGame{AppId: v}
		for i := 0; i < len(game.SellCounts); i++ {
			game.SellCounts[i] = make([]*common.SellCount, days+1)
			for j := 0; j < len(game.SellCounts[i]); j++ {
				game.SellCounts[i][j] = &common.SellCount{
					MapPerson: make(map[string]bool),
				}
			}
		}
		sgr.SellGames = append(sgr.SellGames, game)
	}
	mIds := util.Slice2Map(allIds)
	datas := getSellData(gameId, channel, device, start, end+86399)
	for _, v := range datas {
		indexs := []int{0, mIds[v.miniAppId], mIds[v.miniGameId]}
		ds := []int{days, int(v.ts-start) / 86400}
		for _, index := range indexs {
			for _, d := range ds {
				sgr.SellGames[index].SellCounts[v.event-1][d].SumNum++
				sgr.SellGames[index].SellCounts[v.event-1][d].MapPerson[v.loginname] = true
			}
		}
	}
	mGameNames := GetAllMiniGameNames()
	for _, game := range sgr.SellGames {
		for _, sc := range game.SellCounts {
			for _, v := range sc {
				v.PersonNum = int32(len(v.MapPerson))
			}
		}
		if name := mGameNames[game.AppId]; len(name) > 0 {
			game.AppId += " - " + name
		}
	}
	return sgr
}

func getSellUser(gameId int32, channel, device string, start, end int64) *common.SellUserRes {
	return nil
}

var g_eventLock sync.Mutex
var g_mapEnterGame map[string]int64

var g_awardLock sync.Mutex
var g_mapAward map[string]bool
