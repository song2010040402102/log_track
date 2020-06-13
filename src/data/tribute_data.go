package data

import (
	"bridge"
	"common"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"sort"
	"sync"
	"time"
	"util"
)

type Tribute struct {
	LoginName string `json:"login_name"`
	TS        int64  `json:"ts"`
	Type      int32  `json:"type"`
	Value     int32  `json:"value"`
}

type TributeData struct {
	Parent string     `json:"parent"`
	Childs []*Tribute `json:"childs"`
}

type TriDatas []*TributeData

func (td *TriDatas) AddDatas(datas []*TributeData) {
	if len(*td) == 0 {
		*td = datas
	} else {
		for _, data := range datas {
			exist := false
			for _, cur := range *td {
				if cur.Parent == data.Parent {
					cur.Childs = append(cur.Childs, data.Childs...)
					cur.Childs = util.UniqueSlice2(cur.Childs, func(i, j int) bool {
						return cur.Childs[i].LoginName == cur.Childs[j].LoginName && cur.Childs[i].TS == cur.Childs[j].TS &&
							cur.Childs[i].Type == cur.Childs[j].Type && cur.Childs[i].Value == cur.Childs[j].Value
					}, false).([]*Tribute)
					exist = true
					break
				}
			}
			if !exist {
				*td = append(*td, data)
			}
		}
	}
}

func (td *TriDatas) GetLastTS() int64 {
	ts := int64(0)
	for _, v := range *td {
		for _, vv := range v.Childs {
			if vv.TS > ts {
				ts = vv.TS
			}
		}
	}
	return ts
}

func requestTriDatas(gameId int32, start, end int64) TriDatas {
	var res struct {
		Data  TriDatas `json:"data"`
		Error string   `json:"error"`
	}
	response := bridge.HttpGetServer(gameId, "tribute_pay", fmt.Sprintf("start=%d&end=%d", start, end))
	if err := json.Unmarshal([]byte(response), &res); err != nil {
		logs.Error("requestTriDatas, json unmarshal error:", err)
		return nil
	}
	if res.Error != "" {
		logs.Error("requestTriDatas, response error:", res.Error)
		return nil
	}
	logs.Info("requestTriDatas", gameId, start, end)
	return res.Data
}

func GetTributeDatas(gameId int32, start, end int64, sysType int32) TriDatas {
	g_tdLock.Lock()
	datas := g_mapTribute[gameId]
	g_tdLock.Unlock()
	if last := datas.GetLastTS(); last < end {
		datas.AddDatas(requestTriDatas(gameId, last+1, end))
		g_tdLock.Lock()
		g_mapTribute[gameId] = datas
		g_tdLock.Unlock()
	}
	ret := TriDatas{}
	for _, v := range datas {
		t := &TributeData{}
		t.Parent = v.Parent
		sort.Slice(v.Childs, func(i, j int) bool { return v.Childs[i].TS < v.Childs[j].TS })
		for _, c := range v.Childs {
			if c.TS < start {
				continue
			}
			if c.TS > end {
				break
			}
			if c.Type == sysType {
				t.Childs = append(t.Childs, c)
			}
		}
		if len(t.Childs) > 0 {
			ret = append(ret, t)
		}
	}
	return ret
}

func InitTributeData() {
	g_mapTribute = make(map[int32]TriDatas)
	for _, gameId := range common.GetAllGameId() {
		var datas TriDatas
		util.GobDecodeFile(fmt.Sprintf("data/tribute_%d.dat", gameId), &datas)
		//datas.AddDatas(requestTriDatas(gameId, datas.GetLastTS()+1, time.Now().Unix()))
		g_mapTribute[gameId] = datas
	}
}

func FlushTributeData() {
	for gameId, datas := range g_mapTribute {
		util.GobEncodeFile(fmt.Sprintf("data/tribute_%d.dat", gameId), datas)
	}
}

func AutoGetTributeData() {
	g_tdLock.Lock()
	defer g_tdLock.Unlock()
	for _, gameId := range common.GetAllGameId() {
		datas := g_mapTribute[gameId]
		now, last := time.Now().Unix(), datas.GetLastTS()
		if last < now {
			datas.AddDatas(requestTriDatas(gameId, last+1, now))
			g_mapTribute[gameId] = datas
		}
	}
}

var g_tdLock sync.Mutex
var g_mapTribute map[int32]TriDatas
