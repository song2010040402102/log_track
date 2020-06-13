package data

import (
	"bridge"
	"common"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"sort"
	"sync"
	"util"
)

type UserBriefData struct {
	UUID      int64  `json:"uuid"`
	LoginName string `json:"login_name"`
	RoleName  string `json:"role_name"`
	Channel   int32  `json:"channel"`
	RegTS     int64  `json:"reg_ts"`
	CurRedbag int32  `json:"cur_redbag"`
	SumRedbag int32  `json:"sum_redbag"`
}

func GetUserDatas(gameId int32, loginnames []string) map[string]*UserBriefData {
	ret := make(map[string]*UserBriefData)
	var others []string
	g_ubLock.Lock()
	if mUsers, ok := g_mapUsers[gameId]; ok {
		for _, v := range loginnames {
			if user, ok := mUsers[v]; ok {
				ret[v] = user
			} else {
				others = append(others, v)
			}
		}
	}
	g_ubLock.Unlock()
	if len(others) > 0 {
		sort.Slice(others, func(i, j int) bool { return others[i] < others[j] })
		others = util.UniqueSlice(others, true).([]string)
		var res struct {
			Data  []*UserBriefData `json:"data"`
			Error string           `json:"error"`
		}
		paras := others[0]
		for i := 1; i < len(others); i++ {
			paras += "," + others[i]
		}
		mPara := make(map[string]string)
		mPara["loginnames"] = paras
		response := bridge.HttpGetServer2(gameId, "user_brief", mPara)
		if err := json.Unmarshal([]byte(response), &res); err != nil {
			logs.Error("GetUserDatas, json unmarshal error:", err)
			return ret
		}
		if res.Error != "" {
			logs.Error("GetUserDatas, response error:", res.Error)
			return ret
		}
		g_ubLock.Lock()
		for _, v := range res.Data {
			if v != nil {
				ret[v.LoginName] = v
				g_mapUsers[gameId][v.LoginName] = v
			}
		}
		g_ubLock.Unlock()
	}
	return ret
}

func InitUserData() {
	g_mapUsers = make(map[int32]map[string]*UserBriefData)
	for _, gameId := range common.GetAllGameId() {
		var datas []*UserBriefData
		mUsers := make(map[string]*UserBriefData)
		util.GobDecodeFile(fmt.Sprintf("data/users_%d.dat", gameId), &datas)
		if len(datas) > 0 {
			for _, v := range datas {
				if v != nil {
					mUsers[v.LoginName] = v
				}
			}
		}
		g_mapUsers[gameId] = mUsers
	}
}

func FlushUserData() {
	for gameId, mUsers := range g_mapUsers {
		datas := make([]*UserBriefData, 0, len(mUsers))
		for _, v := range mUsers {
			datas = append(datas, v)
		}
		util.GobEncodeFile(fmt.Sprintf("data/users_%d.dat", gameId), datas)
	}
}

var g_ubLock sync.Mutex
var g_mapUsers map[int32]map[string]*UserBriefData
