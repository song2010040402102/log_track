package common

import (
	"strings"
)

const (
	VERSION_SELL_CFG uint16 = iota + 1
)

const (
	CACHE_TYPE_MYSQL int32 = iota + 1
	CACHE_TYPE_HTTP
)

const (
	MJ_51     int32 = 101 //5151麻将
	MJ_YL     int32 = 102 //运来麻将
	MJ_YL_ZJ  int32 = 103 //麻将赚金
	DDZ_LY    int32 = 201 //乐游斗地主
	DDZ_LQ    int32 = 202 //乐趣斗地主
	DDZ_XM    int32 = 203 //小米斗地主
	DDZ_CS    int32 = 204 //斗地主测试服
	DDZ_WX_CS int32 = 205 //斗地主微信测试服
	DDZ_WX    int32 = 206 //斗地主微信正式服
	XU_DEV    int32 = 901 //旭开发服
)

func GetAllGameId() []int32 {
	return []int32{MJ_51, MJ_YL, MJ_YL_ZJ, DDZ_LY, DDZ_LQ, DDZ_XM, DDZ_CS, DDZ_WX_CS, DDZ_WX}
}

func GetAllServerId() map[int32]string {
	return map[int32]string{
		MJ_51:     "mq-86-020-sh-uqeetest-001",
		MJ_YL:     "mq-86-020-sh-yunlai-001",
		MJ_YL_ZJ:  "mq-86-020-sh-ylzj-000",
		DDZ_LY:    "ddz-86-020-sh-uqeetest-002",
		DDZ_LQ:    "ddz-86-020-sh-uqeetest-003",
		DDZ_XM:    "ddz-86-020-sh-xiaomi-001",
		DDZ_CS:    "ddz-86-020-sh-uqeetest-001",
		DDZ_WX_CS: "ddz-86-020-sh-uqeewx-000",
		DDZ_WX:    "ddz-86-020-sh-uqeewx-001",
	}
}

func GetAllServerName() map[int32]string {
	return map[int32]string{
		MJ_51:     "mj_jinhua",
		MJ_YL:     "mj_jinhua",
		MJ_YL_ZJ:  "mj_jinhua",
		DDZ_LY:    "ddz",
		DDZ_LQ:    "ddz",
		DDZ_XM:    "ddz",
		DDZ_CS:    "ddz",
		DDZ_WX_CS: "ddz",
		DDZ_WX:    "ddz",
	}
}

func GetAllServerIP() map[int32]string {
	return map[int32]string{
		MJ_51:     "s0.mq.uqeetest.uqeegame.cn",
		MJ_YL:     "s1.mq.yunlai.uqeegame.cn",
		MJ_YL_ZJ:  "s1.mq.ylzj.uqeegame.cn",
		DDZ_LY:    "s2.ddz.uqeetest.uqeegame.cn",
		DDZ_LQ:    "s3.ddz.uqeetest.uqeegame.cn",
		DDZ_XM:    "s1.ddz.xiaomi.uqeegame.cn",
		DDZ_CS:    "s1.ddz.uqeetest.uqeegame.cn",
		DDZ_WX_CS: "s0.ddz.wx.uqeegame.cn",
		DDZ_WX:    "s1.ddz.wx.uqeegame.cn",
		XU_DEV:    "10.0.253.27",
	}
}

func GetGameIdByChannel(channel string) int32 {
	m := map[int32]string{
		MJ_51:  "519900",
		DDZ_LY: "515100~515106",
		DDZ_XM: "515107~515199",
		DDZ_CS: "100000",
		DDZ_WX: "519901",
	}
	for k, v := range m {
		if IsMultiCond(v, channel) {
			return k
		}
	}
	return 0
}

type LPItem struct {
	InNum     int32 `json:"in_num"`
	OverNum   int32 `json:"over_num"`
	ClickNum  int32 `json:"click_num"`
	UInNum    int32 `json:"uin_num"`
	UOverNum  int32 `json:"uover_num"`
	UClickNum int32 `json:"uclick_num"`
}

type LPRes struct {
	Ret   int32     `json:"ret"`
	Items []*LPItem `json:"items"`
}

type TransItem struct {
	Click    int32 `json:"click"`
	Active   int32 `json:"active"`
	Register int32 `json:"register"`
}

type TransRes struct {
	Ret   int32        `json:"ret"`
	Items []*TransItem `json:"items"`
}

const (
	SELL_EVENT_APP_ENTER  = 1 //进入小程序
	SELL_EVENT_GAME_ENTER = 2 //进入小游戏
	SELL_EVENT_GAME_EXIT  = 3 //退出小游戏
	SELL_EVENT_END        = 4
)

type SellCount struct {
	SumNum    int32 `json:"sum_num"`
	PersonNum int32 `json:"person_num"`
	MapPerson map[string]bool
}

type SellGame struct {
	AppId      string                           `json:"app_id"`
	SellCounts [SELL_EVENT_END - 1][]*SellCount `json:"sell_counts"`
}

type SellGameRes struct {
	Ret       int32       `json:"ret"`
	SellGames []*SellGame `json:"sell_games"`
}

type SellUserRes struct {
	Ret       int32        `json:"ret"`
	SellUsers [][][]string `json:"sell_users"`
}

func IsMultiCond(cond string, val string) bool {
	if cond == "all" {
		return true
	}
	conds := strings.Split(cond, ",")
	for _, c := range conds {
		cs := strings.Split(c, "~")
		if len(cs) == 2 {
			if val >= cs[0] && val <= cs[1] {
				return true
			}
		} else {
			if c == val {
				return true
			}
		}
	}
	return false
}
