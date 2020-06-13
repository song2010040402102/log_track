package plat

import (
	"common"
	"config"
	"db"
	"fmt"
	"github.com/astaxie/beego/logs"
	"os"
	"sort"
	"strings"
	"sync"
	"util"
	"version"
)

const (
	FLAG_MINI_APP_ID uint32 = 1 << iota
	FLAG_MINI_GAME_ID
	FLAG_MINI_GAME_PATH
	FLAG_MINI_GAME_NAME
	FLAG_MINI_GAME_DETAIL
	FLAG_ICON_URL
	FLAG_ICON_INDEX
	FLAG_CHANNELS
	FLAG_SHOW_POS
	FLAG_PROMOTE_LEVEL
	FLAG_SORT_LEVEL
	FLAG_STAY_TIME
	FLAG_AWARD
)

type SellMiniGameCfg struct {
	MiniAppId      string `json:"mini_app_id"`
	MiniGameId     string `json:"mini_game_id"`
	MiniGamePath   string `json:"mini_game_path"`
	MiniGameName   string `json:"mini_game_name"`
	MiniGameDetail string `json:"mini_game_detail"`
	IconUrl        string `json:"icon_url"`
	IconIndex      uint16 `json:"icon_index"`
	Channels       string `json:"channels"`
	ShowPos        uint64 `json:"show_pos"`
	PromoteLevel   uint8  `json:"promote_level"`
	SortLevel      uint16 `json:"sort_level"`
	StayTime       uint16 `json:"stay_time"`
	Award          uint32 `json:"award"`
}

type SellCfgManager struct {
	Version uint64             `json:"version"`
	SellCfg []*SellMiniGameCfg `json:"sell_cfgs"`
}

type SellMiniGameCfgs []*SellMiniGameCfg

func (s *SellCfgManager) AddCfg(cfg *SellMiniGameCfg) bool {
	if cfg != nil {
		for _, v := range s.SellCfg {
			if v != nil && v.MiniAppId == cfg.MiniAppId && v.MiniGameId == cfg.MiniGameId {
				return false
			}
		}
		cfg.IconUrl = fmt.Sprintf("%s/download/game_%s%s.png", config.Get().FileUrl, cfg.MiniAppId, cfg.MiniGameId)
		s.SellCfg = append(s.SellCfg, cfg)
		s.Version++
	}
	return true
}

func (s *SellCfgManager) ModCfg(cfg *SellMiniGameCfg, flag uint32) bool {
	if cfg != nil {
		for _, v := range s.SellCfg {
			if v != nil && v.MiniAppId == cfg.MiniAppId && v.MiniGameId == cfg.MiniGameId {
				if flag&FLAG_MINI_GAME_PATH != 0 {
					v.MiniGamePath = cfg.MiniGamePath
				}
				if flag&FLAG_MINI_GAME_NAME != 0 {
					v.MiniGameName = cfg.MiniGameName
				}
				if flag&FLAG_MINI_GAME_DETAIL != 0 {
					v.MiniGameDetail = cfg.MiniGameDetail
				}
				if flag&FLAG_ICON_INDEX != 0 {
					v.IconIndex = cfg.IconIndex
				}
				if flag&FLAG_CHANNELS != 0 {
					v.Channels = cfg.Channels
				}
				if flag&FLAG_SHOW_POS != 0 {
					v.ShowPos = cfg.ShowPos
				}
				if flag&FLAG_PROMOTE_LEVEL != 0 {
					v.PromoteLevel = cfg.PromoteLevel
				}
				if flag&FLAG_SORT_LEVEL != 0 {
					v.SortLevel = cfg.SortLevel
				}
				if flag&FLAG_STAY_TIME != 0 {
					v.StayTime = cfg.StayTime
				}
				if flag&FLAG_AWARD != 0 {
					v.Award = cfg.Award
				}
				s.Version++
				return true
			}
		}
	}
	return false
}

func (s *SellCfgManager) DelCfg(miniAppId, miniGameId string) bool {
	for i, v := range s.SellCfg {
		if v != nil && v.MiniAppId == miniAppId && v.MiniGameId == miniGameId {
			s.SellCfg = append(s.SellCfg[:i], s.SellCfg[i+1:]...)
			s.Version++
			return true
		}
	}
	return false
}

func AddSellCfg(cfg *SellMiniGameCfg) bool {
	g_sellLock.Lock()
	defer g_sellLock.Unlock()
	if !g_sellCfgMan.AddCfg(cfg) {
		logs.Error("AddSellCfg, insert failed!")
		return false
	}
	_, err := db.GetMySql().Exec("insert into sell_mini_game_cfg (mini_app_id, mini_game_id, mini_game_path, mini_game_name, mini_game_detail, icon_index, channels, show_pos, promote_level, sort_level, stay_time, award) value(?,?,?,?,?,?,?,?,?,?,?,?)",
		cfg.MiniAppId, cfg.MiniGameId, cfg.MiniGamePath, cfg.MiniGameName, cfg.MiniGameDetail, cfg.IconIndex, cfg.Channels, cfg.ShowPos, cfg.PromoteLevel, cfg.SortLevel, cfg.StayTime, cfg.Award)
	if err != nil {
		logs.Error("AddSellCfg, insert failed with", err)
		return false
	}
	return true
}

func ModSellCfg(cfg *SellMiniGameCfg, flag uint32) bool {
	g_sellLock.Lock()
	defer g_sellLock.Unlock()
	if !g_sellCfgMan.ModCfg(cfg, flag) {
		logs.Error("ModSellCfg, update failed!")
		return false
	}
	var fields string
	if flag&FLAG_MINI_GAME_PATH != 0 {
		fields += fmt.Sprintf(",mini_game_path='%s'", cfg.MiniGamePath)
	}
	if flag&FLAG_MINI_GAME_NAME != 0 {
		fields += fmt.Sprintf(",mini_game_name='%s'", cfg.MiniGameName)
	}
	if flag&FLAG_MINI_GAME_DETAIL != 0 {
		fields += fmt.Sprintf(",mini_game_detail='%s'", cfg.MiniGameDetail)
	}
	if flag&FLAG_ICON_INDEX != 0 {
		fields += fmt.Sprintf(",icon_index=%d", cfg.IconIndex)
	}
	if flag&FLAG_CHANNELS != 0 {
		fields += fmt.Sprintf(",channels='%s'", cfg.Channels)
	}
	if flag&FLAG_SHOW_POS != 0 {
		fields += fmt.Sprintf(",show_pos=%d", cfg.ShowPos)
	}
	if flag&FLAG_PROMOTE_LEVEL != 0 {
		fields += fmt.Sprintf(",promote_level=%d", cfg.PromoteLevel)
	}
	if flag&FLAG_SORT_LEVEL != 0 {
		fields += fmt.Sprintf(",sort_level=%d", cfg.SortLevel)
	}
	if flag&FLAG_STAY_TIME != 0 {
		fields += fmt.Sprintf(",stay_time=%d", cfg.StayTime)
	}
	if flag&FLAG_AWARD != 0 {
		fields += fmt.Sprintf(",award=%d", cfg.Award)
	}
	if fields != "" {
		fields = fields[1:]
		_, err := db.GetMySql().Exec(fmt.Sprintf("update sell_mini_game_cfg set %s where mini_app_id='%s' and mini_game_id='%s'", fields, cfg.MiniAppId, cfg.MiniGameId))
		if err != nil {
			logs.Error("ModSellCfg, update failed with", err)
			return false
		}
	}
	return true
}

func UploadIcon(miniAppId, miniGameId string) string {
	g_sellCfgMan.Version++
	cfg := GetSellCfgById(miniAppId, miniGameId)
	if cfg != nil {
		return cfg.IconUrl
	}
	return ""
}

func DelSellCfg(miniAppId, miniGameId string) bool {
	g_sellLock.Lock()
	defer g_sellLock.Unlock()
	if !g_sellCfgMan.DelCfg(miniAppId, miniGameId) {
		logs.Error("DelSellCfg, delete failed!")
		return false
	}
	_, err := db.GetMySql().Exec("delete from sell_mini_game_cfg where mini_app_id=? and mini_game_id=?", miniAppId, miniGameId)
	if err != nil {
		logs.Error("DelSellCfg, delete failed with", err)
		return false
	}
	os.Remove(fmt.Sprintf("./download/game_%s.png", miniGameId))
	return true
}

func InitSellCfg() {
	rows, err := db.GetMySql().Query("select mini_app_id, mini_game_id, mini_game_path, mini_game_name, mini_game_detail, icon_index, channels, show_pos, promote_level, sort_level, stay_time, award from sell_mini_game_cfg")
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if err != nil {
		logs.Error("InitSellCfg, query failed with %s", err)
		return
	}

	for rows.Next() {
		cfg := &SellMiniGameCfg{}
		err = rows.Scan(&cfg.MiniAppId, &cfg.MiniGameId, &cfg.MiniGamePath, &cfg.MiniGameName, &cfg.MiniGameDetail, &cfg.IconIndex, &cfg.Channels, &cfg.ShowPos, &cfg.PromoteLevel, &cfg.SortLevel, &cfg.StayTime, &cfg.Award)
		if err != nil {
			logs.Error("InitSellCfg, scan failed with %s", err)
			return
		}
		g_sellCfgMan.AddCfg(cfg)
	}
	g_sellCfgMan.Version = version.Get(common.VERSION_SELL_CFG)
}

func SaveSellCfg() {
	version.Set(common.VERSION_SELL_CFG, g_sellCfgMan.Version)
}

func GetSellCfg() []*SellMiniGameCfg {
	g_sellLock.Lock()
	defer g_sellLock.Unlock()
	var ret []*SellMiniGameCfg
	for _, v := range g_sellCfgMan.SellCfg {
		ret = append(ret, v)
	}
	return ret
}

func GetSellCfg2(all bool, channel string) *SellCfgManager {
	g_sellLock.Lock()
	defer g_sellLock.Unlock()
	var sellCfg SellCfgManager
	sellCfg.Version = g_sellCfgMan.Version
	for _, v := range g_sellCfgMan.SellCfg {
		if v == nil {
			continue
		}
		if all || v.PromoteLevel > 0 && common.IsMultiCond(strings.Replace(v.Channels, ";", ",", -1), channel) {
			sellCfg.SellCfg = append(sellCfg.SellCfg, v)
		}
	}
	sort.Slice(sellCfg.SellCfg, func(i, j int) bool { return sellCfg.SellCfg[i].SortLevel < sellCfg.SellCfg[j].SortLevel })
	return &sellCfg
}

func GetSellCfgVer() string {
	return fmt.Sprintf("{\"version\":%d}", g_sellCfgMan.Version)
}

func GetAllMiniAppIds() []string {
	g_sellLock.Lock()
	defer g_sellLock.Unlock()
	var ret []string
	for _, v := range g_sellCfgMan.SellCfg {
		ret = append(ret, v.MiniAppId)
	}
	return util.UniqueSlice(ret, false).([]string)
}

func GetAllMiniGameIds() []string {
	g_sellLock.Lock()
	defer g_sellLock.Unlock()
	var ret []string
	for _, v := range g_sellCfgMan.SellCfg {
		ret = append(ret, v.MiniGameId)
	}
	return util.UniqueSlice(ret, false).([]string)
}

func GetAllMiniGameNames() map[string]string {
	g_sellLock.Lock()
	defer g_sellLock.Unlock()
	m := make(map[string]string)
	for _, v := range g_sellCfgMan.SellCfg {
		m[v.MiniGameId] = v.MiniGameName
	}
	return m
}

func GetSellCfgById(miniAppId, miniGameId string) *SellMiniGameCfg {
	g_sellLock.Lock()
	defer g_sellLock.Unlock()
	for _, v := range g_sellCfgMan.SellCfg {
		if v != nil && v.MiniAppId == miniAppId && v.MiniGameId == miniGameId {
			return v
		}
	}
	return nil
}

func GetSellCfgsByGameId(miniGameId string) []*SellMiniGameCfg {
	g_sellLock.Lock()
	defer g_sellLock.Unlock()
	var ret []*SellMiniGameCfg
	for _, v := range g_sellCfgMan.SellCfg {
		if v != nil && v.MiniGameId == miniGameId {
			ret = append(ret, v)
		}
	}
	return ret
}

var g_sellLock sync.Mutex
var g_sellCfgMan SellCfgManager
