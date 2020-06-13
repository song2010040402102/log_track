package stat

import (
	"data"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
	"util"
)

func GetRelation(gameId int32, item string, startTime, endTime string) *ItemResult {
	switch item {
	case "relation_sum":
		return getRelationSum(gameId, startTime, endTime)
	case "relation_detail":
		return getRelationDetail(gameId, startTime, endTime)
	case "relation_sign_sum":
		return getSignSum(gameId, startTime, endTime)
	case "relation_sign_detail":
		return getSignDetail(gameId, startTime, endTime)
	case "relation_tribute_sum":
		return getTributeSum(gameId, startTime, endTime)
	case "relation_tribute_detail":
		return getTributeDetail(gameId, startTime, endTime)
	}
	return nil
}

type RelationData struct {
	Parent string
	Childs []string
}

type Relations []*RelationData

func (r *Relations) AddRelation(parent string, child string) {
	for _, v := range *r {
		if v.Parent == parent {
			v.Childs = append(v.Childs, child)
			return
		}
	}
	*r = append(*r, &RelationData{parent, []string{child}})
}

func getRelationDatas(gameId int32, start, end int64) Relations {
	var r Relations
	err := data.GetLogData(gameId, "WeChatShareLoginFlow",
		util.Ts2date(start-start%86400), util.Ts2date(end-end%86400), func(day int32, pos int32, s []string) bool {
			if pos <= 0 {
				return true
			}
			if len(s) < 11 {
				logs.Error("getRelationDatas data error!")
				return false
			}
			ts, _ := strconv.ParseInt(s[6], 10, 32)
			if ts >= start && ts <= end {
				r.AddRelation(s[10], s[0])
			}
			return true
		})
	if err != nil {
		return nil
	}
	return r
}

func getRelationSum(gameId int32, startTime, endTime string) *ItemResult {
	var loginnames []string
	relDatas := getRelationDatas(gameId, util.Time2ts(startTime), util.Time2ts(endTime))
	for _, v := range relDatas {
		if v == nil {
			continue
		}
		loginnames = append(loginnames, v.Parent)
	}
	mUsers := data.GetUserDatas(gameId, loginnames)
	heads := []string{"登录名", "角色名", "渠道号", "注册时间", "徒弟数"}
	items := make([][]string, 0, len(relDatas))
	for _, v := range relDatas {
		if v == nil {
			continue
		}
		rows := make([]string, 0, len(heads))
		rows = append(rows, v.Parent)
		if user, _ := mUsers[v.Parent]; user != nil {
			rows = append(rows, user.RoleName)
			rows = append(rows, strconv.Itoa(int(user.Channel)))
			rows = append(rows, util.Ts2time(user.RegTS))
		} else {
			rows = append(rows, "")
			rows = append(rows, "")
			rows = append(rows, "")
		}
		rows = append(rows, strconv.Itoa(len(v.Childs)))
		items = append(items, rows)
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("师徒数据汇总(%s ~ %s)", startTime, endTime), heads, items),
	}
}

func getRelationDetail(gameId int32, startTime, endTime string) *ItemResult {
	var loginnames []string
	relDatas := getRelationDatas(gameId, util.Time2ts(startTime), util.Time2ts(endTime))
	for _, v := range relDatas {
		if v == nil {
			continue
		}
		loginnames = append(loginnames, v.Parent)
		for _, child := range v.Childs {
			loginnames = append(loginnames, child)
		}
	}
	mUsers := data.GetUserDatas(gameId, loginnames)
	heads := []string{"师傅登录名", "师傅角色名", "师傅渠道号", "师傅注册时间", "徒弟登录名", "徒弟角色名", "徒弟渠道号", "徒弟注册时间"}
	items := make([][]string, 0, len(relDatas))
	for _, v := range relDatas {
		if v == nil {
			continue
		}
		for i, child := range v.Childs {
			rows := make([]string, 0, len(heads))
			if i == 0 {
				rows = append(rows, v.Parent)
				if user, _ := mUsers[v.Parent]; user != nil {
					rows = append(rows, user.RoleName)
					rows = append(rows, strconv.Itoa(int(user.Channel)))
					rows = append(rows, util.Ts2time(user.RegTS))
				} else {
					rows = append(rows, "")
					rows = append(rows, "")
					rows = append(rows, "")
				}
			} else {
				rows = append(rows, "")
				rows = append(rows, "")
				rows = append(rows, "")
				rows = append(rows, "")
			}
			rows = append(rows, child)
			if user, _ := mUsers[child]; user != nil {
				rows = append(rows, user.RoleName)
				rows = append(rows, strconv.Itoa(int(user.Channel)))
				rows = append(rows, util.Ts2time(user.RegTS))
			} else {
				rows = append(rows, "")
				rows = append(rows, "")
				rows = append(rows, "")
			}
			items = append(items, rows)
		}
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("师徒数据明细(%s ~ %s)", startTime, endTime), heads, items),
	}
}

func getSignSum(gameId int32, startTime, endTime string) *ItemResult {
	var loginnames []string
	triDatas := data.GetTributeDatas(gameId, util.Time2ts(startTime), util.Time2ts(endTime), 65)
	for _, v := range triDatas {
		if v == nil {
			continue
		}
		loginnames = append(loginnames, v.Parent)
	}
	mUsers := data.GetUserDatas(gameId, loginnames)
	heads := []string{"登录名", "角色名", "渠道号", "注册时间", "徒弟签到个数"}
	items := make([][]string, 0, len(triDatas))
	for _, v := range triDatas {
		if v == nil {
			continue
		}
		rows := make([]string, 0, len(heads))
		rows = append(rows, v.Parent)
		if user, _ := mUsers[v.Parent]; user != nil {
			rows = append(rows, user.RoleName)
			rows = append(rows, strconv.Itoa(int(user.Channel)))
			rows = append(rows, util.Ts2time(user.RegTS))
		} else {
			rows = append(rows, "")
			rows = append(rows, "")
			rows = append(rows, "")
		}
		rows = append(rows, strconv.Itoa(len(v.Childs)))
		items = append(items, rows)
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("徒弟签到汇总(%s ~ %s)", startTime, endTime), heads, items),
	}
}

func getSignDetail(gameId int32, startTime, endTime string) *ItemResult {
	var loginnames []string
	triDatas := data.GetTributeDatas(gameId, util.Time2ts(startTime), util.Time2ts(endTime), 65)
	for _, v := range triDatas {
		if v == nil {
			continue
		}
		for _, child := range v.Childs {
			loginnames = append(loginnames, child.LoginName)
		}
	}
	mUsers := data.GetUserDatas(gameId, loginnames)
	heads := []string{"登录名", "角色名", "渠道号", "注册时间", "签到第1天", "签到第2天", "签到第3天", "签到第4天", "签到第5天"}
	items := make([][]string, 0, len(triDatas))
	for _, v := range triDatas {
		if v == nil || len(v.Childs) == 0 {
			continue
		}
		rows := make([]string, 0, len(heads))
		mSign := make(map[string][]int64)
		for _, child := range v.Childs {
			mSign[child.LoginName] = append(mSign[child.LoginName], child.TS)
		}
		for loginname, tss := range mSign {
			rows = append(rows, loginname)
			if user, _ := mUsers[loginname]; user != nil {
				rows = append(rows, user.RoleName)
				rows = append(rows, strconv.Itoa(int(user.Channel)))
				rows = append(rows, util.Ts2time(user.RegTS))
			} else {
				rows = append(rows, "")
				rows = append(rows, "")
				rows = append(rows, "")
			}
			for _, ts := range tss {
				rows = append(rows, util.Ts2time(ts))
			}
			for i := 0; i < 5-len(tss); i++ {
				rows = append(rows, "")
			}
		}
		items = append(items, rows)
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("徒弟签到明细(%s ~ %s)", startTime, endTime), heads, items),
	}
}

func getTributeSum(gameId int32, startTime, endTime string) *ItemResult {
	var loginnames []string
	triDatas := data.GetTributeDatas(gameId, util.Time2ts(startTime), util.Time2ts(endTime), 63)
	for _, v := range triDatas {
		if v == nil {
			continue
		}
		loginnames = append(loginnames, v.Parent)
	}
	mUsers := data.GetUserDatas(gameId, loginnames)
	heads := []string{"登录名", "角色名", "渠道号", "注册时间", "徒弟进贡红包券"}
	items := make([][]string, 0, len(triDatas))
	for _, v := range triDatas {
		if v == nil {
			continue
		}
		rows := make([]string, 0, len(heads))
		rows = append(rows, v.Parent)
		if user, _ := mUsers[v.Parent]; user != nil {
			rows = append(rows, user.RoleName)
			rows = append(rows, strconv.Itoa(int(user.Channel)))
			rows = append(rows, util.Ts2time(user.RegTS))
		} else {
			rows = append(rows, "")
			rows = append(rows, "")
			rows = append(rows, "")
		}
		sum := int32(0)
		for _, child := range v.Childs {
			sum += child.Value
		}
		rows = append(rows, strconv.Itoa(int(sum)))
		items = append(items, rows)
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("徒弟进贡汇总(%s ~ %s)", startTime, endTime), heads, items),
	}
}

func getTributeDetail(gameId int32, startTime, endTime string) *ItemResult {
	var loginnames []string
	triDatas := data.GetTributeDatas(gameId, util.Time2ts(startTime), util.Time2ts(endTime), 63)
	for _, v := range triDatas {
		if v == nil {
			continue
		}
		loginnames = append(loginnames, v.Parent)
		for _, child := range v.Childs {
			loginnames = append(loginnames, child.LoginName)
		}
	}
	mUsers := data.GetUserDatas(gameId, loginnames)
	heads := []string{"师傅登录名", "师傅角色名", "师傅渠道号", "师傅注册时间", "时间", "进贡红包券", "徒弟登录名", "徒弟角色名", "徒弟渠道号", "徒弟注册时间"}
	items := make([][]string, 0, len(triDatas))
	for _, v := range triDatas {
		if v == nil {
			continue
		}
		for i, child := range v.Childs {
			rows := make([]string, 0, len(heads))
			if i == 0 {
				rows = append(rows, v.Parent)
				if user, _ := mUsers[v.Parent]; user != nil {
					rows = append(rows, user.RoleName)
					rows = append(rows, strconv.Itoa(int(user.Channel)))
					rows = append(rows, util.Ts2time(user.RegTS))
				} else {
					rows = append(rows, "")
					rows = append(rows, "")
					rows = append(rows, "")
				}
			} else {
				rows = append(rows, "")
				rows = append(rows, "")
				rows = append(rows, "")
				rows = append(rows, "")
			}
			rows = append(rows, util.Ts2time(child.TS))
			rows = append(rows, strconv.Itoa(int(child.Value)))
			rows = append(rows, child.LoginName)
			if user, _ := mUsers[child.LoginName]; user != nil {
				rows = append(rows, user.RoleName)
				rows = append(rows, strconv.Itoa(int(user.Channel)))
				rows = append(rows, util.Ts2time(user.RegTS))
			} else {
				rows = append(rows, "")
				rows = append(rows, "")
				rows = append(rows, "")
			}
			items = append(items, rows)
		}
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("徒弟进贡明细(%s ~ %s)", startTime, endTime), heads, items),
	}
}
