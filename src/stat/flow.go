package stat

import (
	"data"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
	"strings"
	"util"
)

var g_mapFlowTitle map[string][]string = map[string][]string{
	"User":                  []string{"uuid", "createTime", "loginName", "roleName", "channel", "curRedbag-当前红包券", "sumRedbag-累计获取红包券", "dedRedbag-累计扣除红包券", "sumIncome-累计提现(元)", "sumTelebill-累计提话费(元)", "sumPay-累计充值(元)"},
	"WxCreate":              []string{"scene", "source", "loginCode", "ts"},
	"Create":                []string{"clientType"},
	"Login":                 []string{"offlineTime", "registerTime"},
	"Logout":                []string{"onlineTime", "registerTime"},
	"Online":                []string{"timeDate", "count"},
	"AdToMinGameFlow":       []string{"appid", "where"},
	"AdToMinGameResultFlow": []string{"appid", "where"},
	"DiamondFlow":           []string{"ruleType", "pre", "cur", "delta", "where", "flowType", "extra", "extra1"},
	"GoldFlow":              []string{"ruleType", "pre", "cur", "delta", "where", "flowType", "extra", "extra1"},
	"FangkaFlow":            []string{"ruleType", "pre", "cur", "delta", "where", "flowType"},
	"VoucherFlow":           []string{"ruleType", "pre", "cur", "delta", "where", "flowType"},
	"EnrollVoucherFlow":     []string{"ruleType", "pre", "cur", "delta", "where", "flowType"},
	"StarIntegalFlow":       []string{"ruleType", "pre", "cur", "delta", "where", "flowType", "extra", "extra1"},
	"ZjjIntegalFlow":        []string{"ruleType", "pre", "cur", "delta", "where", "flowType", "extra", "extra1"},
	"TableStatFlow":         []string{"ruleType", "pre", "cur", "delta", "where", "flowType", "extra", "extra1"},
	"VipExpFlow":            []string{"ruleType", "pre", "cur", "delta", "where", "flowType"},
	"WheelFlow":             []string{"ruleType", "pre", "cur", "delta", "where", "flowType"},
	"MaterialFlow":          []string{"ruleType", "pre", "cur", "delta", "where", "flowType", "id"},
	"RedbagConsumeFlow":     []string{"ruleType", "pre", "cur", "delta", "where", "flowType"},
	"TeleBillFlow":          []string{"ruleType", "pre", "cur", "delta", "where", "flowType", "extra", "extra1"},
	"TaskFlow":              []string{"taskType", "taskId", "nums", "progress"},
	"TaskFinishFlow":        []string{"taskType", "taskId"},
	"TaskDrawFlow":          []string{"taskType", "taskId"},
	"PayFlow":               []string{"cash", "diamond", "gold", "fangka", "bonus", "chargeDiamond", "orderId", "isFirst", "comment"},
	"IncomeFlow":            []string{"uid", "cash"},
	"IncomeTelebillFlow":    []string{"uid", "cash"},
	"RedBagFlow":            []string{"cash", "where", "extra", "extra1"},
	"DrawCashFlow":          []string{"cash", "tradeNo", "wxOrderNo"},
	"RoomFlow":              []string{"roomId", "ruleType", "roomType", "roomLevel"},
	"RealRoomFlow":          []string{"roomId", "ruleType", "roomType", "roomLevel", "round", "maxRound"},
	"RoomNoBenifitFlow":     []string{"roomId", "ruleType", "roomType", "roomLevel"},
	"RoomMergePlayFlow":     []string{"roomId", "ruleType", "roomType", "roomLevel", "pre", "cur", "delta", "flowType"},
	"RoomAutoFlow":          []string{"roomId", "ruleType", "roomType", "roomLevel", "autoTimes"},
	"MatchFlow":             []string{"matchId", "matchXmlId", "ruleType", "numsLimit"},
	"MatchResultFlow":       []string{"matchId", "matchXmlId", "ruleType", "numsLimit", "rank"},
	"GrandPrixFlow":         []string{"matchId", "matchXmlId", "ruleType", "numsLimit"},
	"GrandPrixResultFlow":   []string{"matchId", "matchXmlId", "ruleType", "numsLimit", "rank"},
	"VideoStartFlow":        []string{"videoType", "where"},
	"VideoEndFlow":          []string{"videoType", "where"},
	"VideoClickFlow":        []string{"videoType", "where"},
	"VideoLoginFlow":        []string{"videoType", "where"},
	"VideoLoginClickFlow":   []string{"videoType", "where"},
	"VideoInsertFlow":       []string{"videoType", "where"},
	"VideoInsertClickFlow":  []string{"videoType", "where"},
	"BannerStartFlow":       []string{"videoType", "where"},
	"BannerEndFlow":         []string{"videoType", "where"},
	"BannerClickFlow":       []string{"videoType", "where"},
	"WeChatSharePicFlow":    []string{"picId", "inviterloginName", "isNew"},
	"WeChatShareClickFlow":  []string{"first", "unionid", "nickname"},
	"WeChatShareLoginFlow":  []string{"inviterLoginName"},
}

var g_mapFlowNoHead map[string]bool = map[string]bool{
	"WxCreate":     true,
	"Online":       true,
	"RealRoomFlow": true,
}

func IsCustomFlow(flow string) bool {
	if flow == "User" {
		return true
	}
	return false
}

func GetCustomFlow(gameId int32, flow string, startTime, endTime string, fCond func([]string) bool) [][]string {
	if flow == "User" {
		return GetUserFlow(gameId, flow, startTime, endTime, fCond)
	}
	return nil
}

func GetUserFlow(gameId int32, flow string, startTime, endTime string, fCond func([]string) bool) [][]string {
	loginnames := GetUserLoginnames(gameId, startTime, endTime)
	if len(loginnames) == 0 {
		return nil
	}
	mUsers := data.GetUserDatas(gameId, loginnames)
	var mCashs [3]map[string]int64
	flows := []string{"IncomeFlow", "IncomeTelebillFlow", "PayFlow"}
	for i, flow := range flows {
		mCashs[i] = make(map[string]int64)
		data.GetLogData(gameId, flow, startTime[:10], util.GetDate(), func(day int32, pos int32, s []string) bool {
			if pos <= 0 {
				return true
			}
			if len(s) < 12 {
				logs.Error(flow, "data error!")
				return false
			}
			cash, _ := strconv.ParseInt(s[11], 10, 32)
			mCashs[i][s[0]] += cash
			return true
		})
	}
	var ret [][]string
	for _, loginname := range loginnames {
		s := make([]string, len(g_mapFlowTitle["User"]))
		s[2] = loginname
		if user, _ := mUsers[loginname]; user != nil {
			s[0] = strconv.FormatInt(user.UUID, 10)
			s[1] = util.Ts2time(user.RegTS)
			s[3] = user.RoleName
			s[4] = strconv.Itoa(int(user.Channel))
			s[5] = strconv.Itoa(int(user.CurRedbag))
			s[6] = strconv.Itoa(int(user.SumRedbag))
			s[7] = strconv.Itoa(int(user.SumRedbag - user.CurRedbag))
		}
		for i, cash := range mCashs {
			if i == 0 || i == 1 {
				s[8+i] = strconv.FormatFloat(float64(cash[loginname])/float64(GetCashScaleByGameId(gameId)), 'f', 2, 64)
			} else {
				s[8+i] = strconv.FormatInt(cash[loginname], 10)
			}
		}
		if fCond(s) {
			ret = append(ret, s)
		}
	}
	return ret
}

func GetUserLoginnames(gameId int32, startTime, endTime string) []string {
	var loginnames []string
	data.GetLogData(gameId, "Create", startTime[:10], endTime[:10], func(day int32, pos int32, s []string) bool {
		if pos <= 0 {
			return true
		}
		if len(s) < 7 {
			logs.Error("Create data error!")
			return false
		}
		ts, _ := strconv.ParseInt(s[6], 10, 32)
		if util.Time2ts(startTime) <= ts && ts <= util.Time2ts(endTime) {
			loginnames = append(loginnames, s[0])
		}
		return true
	})
	return loginnames
}

func ShowFlow(gameId int32, flow string, startTime, endTime string, roleName string, filter string) *ItemResult {
	startTs := util.Time2ts(startTime)
	endTs := util.Time2ts(endTime)
	if startTs > endTs {
		return nil
	}

	var conds [][]*ConditionInfo
	if filter != "" {
		conds = parseFilter(filter)
		if conds == nil || len(conds) == 0 {
			return &ItemResult{
				Ret: 1,
				Msg: fmt.Sprintf("%s filter syntax error!", filter),
			}
		}
	}

	heads := GetHead(flow)
	keys := make([]string, len(heads))
	for i, v := range heads {
		ss := strings.Split(v, "-")
		keys[i] = ss[0]
	}
	for _, v := range conds {
		for _, vv := range v {
			index := GetFieldIndex(keys, vv.key)
			if index == -1 {
				return &ItemResult{
					Ret: 2,
					Msg: fmt.Sprintf("%s key not exist!", vv.key),
				}
			}
		}
	}

	fCond := func(s []string) bool {
		if roleName != "" {
			iRole := GetFieldIndex(keys, "roleName")
			if iRole < 0 || iRole >= len(s) || s[iRole] != roleName {
				return false
			}
		}
		if len(conds) != 0 && !isCondition(keys, s, conds) {
			return false
		}
		return true
	}

	var items [][]string
	if IsCustomFlow(flow) {
		items = GetCustomFlow(gameId, flow, startTime, endTime, fCond)
	} else {
		err := data.GetLogData(gameId, flow, startTime[:10], endTime[:10], func(day int32, pos int32, s []string) bool {
			if pos <= 0 {
				return true
			}
			if len(s) < 2 {
				logs.Error("ShowFlow data error!")
				return false
			}
			if ts := util.Time2ts(s[1]); ts < startTs || ts > endTs {
				return true
			}
			if fCond(s) {
				items = append(items, s)
			}
			return true
		})
		if err != nil {
			return &ItemResult{
				Ret: 3,
				Msg: err.Error(),
			}
		}
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("%s(%s ~ %s)", flow, startTime, endTime), heads, items),
	}
}

func GetHead(flow string) []string {
	if IsCustomFlow(flow) {
		return g_mapFlowTitle[flow]
	}
	heads := []string{"loginName", "time"}
	if _, ok := g_mapFlowNoHead[flow]; !ok {
		heads = append(heads, []string{"uid", "roleName", "platform", "unionId", "regTime", "regIP", "channel", "device"}...)
	}
	return append(heads, g_mapFlowTitle[flow]...)
}

func GetFieldIndex(heads []string, field string) int {
	field = strings.ToLower(field)
	for i, v := range heads {
		if field == strings.ToLower(v) {
			return i
		}
	}
	return -1
}

const (
	REL_GE int8 = iota // >=
	REL_GT             // >
	REL_LE             // <=
	REL_LT             // <
	REL_NE             // !=
	REL_EQ             // =
)

const (
	VT_NONE       int8 = iota //没有值类型
	VT_STRING_SET             //字符串集合
	VT_FLOAT_SEC              //浮点区间
	VT_FLOAT_SEC1             //浮点区间[]
	VT_FLOAT_SEC2             //浮点区间[)
	VT_FLOAT_SEC3             //浮点区间(]
	VT_FLOAT_SEC4             //浮点区间()
)

type ConditionInfo struct {
	rel       int8
	key       string
	valueType int8
	values    []string
}

func parseFilter(filter string) [][]*ConditionInfo {
	filter = strings.Replace(filter, " ", "", -1)
	filter = strings.Replace(filter, "	", "", -1)
	orF := strings.Split(filter, "||")
	if len(orF) == 0 {
		return nil
	}
	orRet := make([][]*ConditionInfo, 0, len(orF))
	for _, of := range orF {
		andF := strings.Split(of, "&&")
		if len(andF) == 0 {
			return nil
		}
		andRet := make([]*ConditionInfo, 0, len(andF))
		rel := []string{">=", ">", "<=", "<", "!=", "="}
		for _, af := range andF {
			cond := &ConditionInfo{}
			for i, r := range rel {
				kv := strings.Split(af, r)
				if len(kv) != 2 {
					continue
				}
				if len(kv[0]) == 0 || len(kv[1]) == 0 {
					return nil
				}
				cond.rel = int8(i)
				cond.key = kv[0]
				if i >= 0 && i <= 3 { //单值区间
					if _, err := strconv.ParseFloat(kv[1], 32); err != nil {
						return nil
					}
					cond.values = append(cond.values, kv[1])
				} else if kv[1][0] == '[' || kv[1][0] == '(' { //双值区间
					if len(kv[1]) < 5 {
						return nil
					}
					sv := strings.Split(kv[1][1:len(kv[1])-1], ",")
					if len(sv) != 2 {
						return nil
					}
					for _, v := range sv {
						if _, err := strconv.ParseFloat(v, 32); err == nil {
							cond.valueType = VT_FLOAT_SEC
						} else {
							return nil
						}
					}
					if kv[1][0] == '[' {
						if kv[1][len(kv[1])-1] == ']' {
							cond.valueType += 1
						} else if kv[1][len(kv[1])-1] == ')' {
							cond.valueType += 2
						} else {
							return nil
						}
					} else if kv[1][0] == '(' {
						if kv[1][len(kv[1])-1] == ']' {
							cond.valueType += 3
						} else if kv[1][len(kv[1])-1] == ')' {
							cond.valueType += 4
						} else {
							return nil
						}
					} else {
						return nil
					}
					cond.values = append(cond.values, sv...)
				} else { //集合
					sv := strings.Split(kv[1], ",")
					cond.valueType = VT_STRING_SET
					cond.values = append(cond.values, sv...)
				}
				break
			}
			andRet = append(andRet, cond)
		}
		orRet = append(orRet, andRet)
	}
	return orRet
}

func isCondition(fields []string, values []string, conds [][]*ConditionInfo) bool {
	for _, orc := range conds {
		andCond := true
		for _, cond := range orc {
			index := GetFieldIndex(fields, cond.key)
			if index < 0 || index >= len(values) {
				return false
			}
			if cond.rel >= REL_GE && cond.rel <= REL_LT { //单值区间判断
				v1, _ := strconv.ParseFloat(values[index], 32)
				v2, _ := strconv.ParseFloat(cond.values[0], 32)
				if cond.rel == REL_GE {
					if v1 < v2 {
						andCond = false
					}
				} else if cond.rel == REL_GT {
					if v1 <= v2 {
						andCond = false
					}
				} else if cond.rel == REL_LE {
					if v1 > v2 {
						andCond = false
					}
				} else {
					if v1 >= v2 {
						andCond = false
					}
				}
			} else if cond.valueType >= VT_FLOAT_SEC1 && cond.valueType <= VT_FLOAT_SEC4 { //双值区间判断
				v1, _ := strconv.ParseFloat(values[index], 32)
				v2, _ := strconv.ParseFloat(cond.values[0], 32)
				v3, _ := strconv.ParseFloat(cond.values[1], 32)
				if cond.valueType == VT_FLOAT_SEC1 {
					if v1 >= v2 && v1 <= v3 {
						if cond.rel == REL_NE {
							andCond = false
						}
					} else if cond.rel == REL_EQ {
						andCond = false
					}
				} else if cond.valueType == VT_FLOAT_SEC2 {
					if v1 >= v2 && v1 < v3 {
						if cond.rel == REL_NE {
							andCond = false
						}
					} else if cond.rel == REL_EQ {
						andCond = false
					}
				} else if cond.valueType == VT_FLOAT_SEC3 {
					if v1 > v2 && v1 <= v3 {
						if cond.rel == REL_NE {
							andCond = false
						}
					} else if cond.rel == REL_EQ {
						andCond = false
					}
				} else {
					if v1 > v2 && v1 < v3 {
						if cond.rel == REL_NE {
							andCond = false
						}
					} else if cond.rel == REL_EQ {
						andCond = false
					}
				}
			} else { //集合判断
				if cond.rel == REL_EQ {
					exist := false
					for _, v := range cond.values {
						if v == values[index] {
							exist = true
							break
						}
					}
					if !exist {
						andCond = false
					}
				} else {
					for _, v := range cond.values {
						if v == values[index] {
							andCond = false
							break
						}
					}
				}
			}
			if !andCond {
				break
			}
		}
		if andCond {
			return true
		}
	}
	return false
}
