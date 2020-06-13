package stat

import (
	"fmt"
	//"github.com/astaxie/beego/logs"
	"common"
	"data"
	"sort"
	"strconv"
	"util"
)

func AdvertisingHandler(gameId int32, startDate, endDate string, channel string, device string) *ItemResult {
	adInfo := getAdvertising(gameId, startDate, endDate, channel, device)
	plats, vtypes := getPlatVTypes(adInfo)
	heads := make([]string, 0, len(vtypes)+1)
	heads = append(heads, "")
	for _, vtype := range vtypes {
		heads = append(heads, vtype)
	}
	childs := make([][]*TreeTable, 0, len(plats))
	for _, plat := range plats {
		childx := make([]*TreeTable, 0, len(vtypes)+1)
		childx = append(childx, NewTable(data.PLAT_NAME[plat], nil, nil))
		for _, vtype := range vtypes {
			heads := []string{"日期", "渠道", "设备", "展现量", "点击量", "点击率", "CPM", "预估收益(元)"}
			items := make([][]string, 0, len(adInfo[plat][vtype]))
			for i, v := range adInfo[plat][vtype] {
				rows := make([]string, 0, len(heads))
				rows = append(rows, util.Ts2date(util.Date2ts(startDate)+int64(i*86400)))
				rows = append(rows, channel)
				rows = append(rows, device)
				rows = append(rows, strconv.Itoa(int(v.Show)))
				rows = append(rows, strconv.Itoa(int(v.Click)))
				if v.Show > 0 {
					rows = append(rows, fmt.Sprintf("%.2f%%", float32(v.Click)*100/float32(v.Show)))
					rows = append(rows, strconv.FormatFloat(float64(v.Income)*1000/float64(v.Show), 'f', 2, 64))
				} else {
					rows = append(rows, "0.00%")
					rows = append(rows, "0.00")
				}
				rows = append(rows, strconv.FormatFloat(v.Income, 'f', 2, 64))
				items = append(items, rows)
			}
			childx = append(childx, NewTable("", heads, items))
		}
		childs = append(childs, childx)
	}
	return &ItemResult{
		TTable: NewTreeTable(fmt.Sprintf("广告统计(%s ~ %s)", startDate, endDate), heads, childs),
	}
}

func getAdvertising(gameId int32, startDate, endDate string, channel string, device string) map[int32]map[string][]*data.ADData {
	adIds := []string{}
	for _, v := range data.GetAllAdInfo() {
		if gameId == v.GameId && common.IsMultiCond(channel, v.Channel) && common.IsMultiCond(device, v.Device) {
			adIds = append(adIds, v.AdId)
		}
	}
	mPlat := map[int32]bool{data.AD_ALL: true}
	mVType := map[string]bool{"all": true}
	for _, v := range adIds {
		adInfo := data.GetAdInfo(v)
		if adInfo != nil {
			mPlat[adInfo.PlatId] = true
			mVType[adInfo.VType] = true
		}
	}
	start, end := util.Date2ts(startDate), util.Date2ts(endDate)
	days := int((end-start)/86400 + 1)

	ret := make(map[int32]map[string][]*data.ADData)
	for plat, _ := range mPlat {
		mType := make(map[string][]*data.ADData)
		for vtype, _ := range mVType {
			mType[vtype] = make([]*data.ADData, days)
			for i := 0; i < days; i++ {
				mType[vtype][i] = &data.ADData{}
			}
		}
		ret[plat] = mType
	}

	adDatas := data.GetAdDatas(start, end, adIds)
	for plat, _ := range mPlat {
		for vtype, _ := range mVType {
			for i := 0; i < days; i++ {
				indexs := getAdDataIndexs(adDatas, plat, vtype, start+int64(i*86400))
				for _, index := range indexs {
					ret[plat][vtype][i].Show += adDatas[index].Show
					ret[plat][vtype][i].Click += adDatas[index].Click
					ret[plat][vtype][i].Income += adDatas[index].Income
				}
			}
		}
	}
	return ret
}

func getAdDataIndexs(adDatas []*data.ADData, plat int32, vtype string, ts int64) []int {
	indexs := []int{}
	for i, ad := range adDatas {
		if ad.TS == ts {
			adInfo := data.GetAdInfo(ad.AdId)
			if adInfo != nil && (plat == data.AD_ALL || plat == adInfo.PlatId) && (vtype == "all" || vtype == adInfo.VType) {
				indexs = append(indexs, i)
			}
		}
	}
	return indexs
}

func getPlatVTypes(adInfo map[int32]map[string][]*data.ADData) ([]int32, []string) {
	if len(adInfo) == 0 {
		return nil, nil
	}
	plats := []int32{}
	for k, _ := range adInfo {
		plats = append(plats, k)
	}
	sort.Slice(plats, func(i, j int) bool { return plats[i] < plats[j] })
	vtypes := []string{}
	for k, _ := range adInfo[0] {
		vtypes = append(vtypes, k)
	}
	sort.Slice(vtypes, func(i, j int) bool {
		if vtypes[i] == "all" {
			return true
		} else if vtypes[j] == "all" {
			return false
		} else {
			return vtypes[i] < vtypes[j]
		}
	})
	return plats, vtypes
}
