package data

import (
	"bufio"
	"common"
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/astaxie/beego/logs"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"util"
)

const (
	AD_ALL int32 = 0
	AD_CSJ int32 = 101
	AD_YLH int32 = 102
	AD_XM  int32 = 103
)

var PLAT_NAME map[int32]string = map[int32]string{
	AD_ALL: "all",
	AD_CSJ: "穿山甲",
	AD_YLH: "优量汇",
	AD_XM:  "小米广告",
}

type ADInfo struct {
	GameId  int32
	PlatId  int32
	Device  string
	Channel string
	VType   string
	AdId    string
}

var g_adInfos []*ADInfo = []*ADInfo{
	&ADInfo{common.DDZ_LY, AD_CSJ, "ios", "515100", "开屏", "824556468"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "ios", "515100", "Banner", "924556237"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "ios", "515100", "插屏", "924556756"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "ios", "515100", "横版激励视频", "924556415"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "ios", "515100", "竖版激励视频", "924556461"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "ios", "515101", "开屏", "824556183"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "ios", "515101", "Banner", "924556988"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "ios", "515101", "插屏", "924556517"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "ios", "515101", "横版激励视频", "924556009"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "ios", "515101", "竖版激励视频", "924556869"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515100", "开屏", "826063853"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515100", "Banner", "926063981"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515100", "插屏", "926063634"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515100", "横版激励视频", "926063817"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515100", "竖版激励视频", "926063505"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515101", "开屏", "826063757"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515101", "Banner", "926063950"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515101", "插屏", "926063036"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515101", "横版激励视频", "926063869"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515101", "竖版激励视频", "926063574"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515103", "开屏", "826063757"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515103", "Banner", "926063950"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515103", "插屏", "926063036"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515103", "横版激励视频", "926063869"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515103", "竖版激励视频", "926063574"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515188", "开屏", "826063905"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515188", "Banner", "926063905"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515188", "插屏", "926063143"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515188", "横版激励视频", "926063722"},
	&ADInfo{common.DDZ_LY, AD_CSJ, "android", "515188", "竖版激励视频", "926063415"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515100", "开屏", "4070581472622313"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515100", "Banner", "2000987452626452"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515100", "插屏", "9010384492021413"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515100", "横版激励视频", "3020187472929592"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515100", "竖版激励视频", "3020187472929592"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515101", "开屏", "1070186442925513"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515101", "Banner", "8040980442927574"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515101", "插屏", "2000380412122565"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515101", "横版激励视频", "4040187492325586"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515101", "竖版激励视频", "4040187492325586"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515103", "开屏", "1070186442925513"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515103", "Banner", "8040980442927574"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515103", "插屏", "2000380412122565"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515103", "横版激励视频", "4040187492325586"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515103", "竖版激励视频", "4040187492325586"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515188", "开屏", "2000987448944162"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515188", "Banner", "2030189468448123"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515188", "插屏", "4090684418445174"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515188", "横版激励视频", "6050582428641242"},
	&ADInfo{common.DDZ_LY, AD_YLH, "android", "515188", "竖版激励视频", "6050582428641242"},
	&ADInfo{common.DDZ_LY, AD_XM, "android", "515189", "开屏", "fd49e1364116200909e58dae5e07ebb2"},
	&ADInfo{common.DDZ_LY, AD_XM, "android", "515189", "Banner", "52a299ed5445f3b05ac79563a8d376b4"},
	&ADInfo{common.DDZ_LY, AD_XM, "android", "515189", "插屏", "2f4bcafdc8dee863aae1e3acbe6dbe8a"},
	&ADInfo{common.DDZ_LY, AD_XM, "android", "515189", "横版激励视频", "01eb107f923908d528bd6ab57428b7b9"},
	&ADInfo{common.DDZ_LY, AD_XM, "android", "515189", "竖版激励视频", "cc4c69d7d8c3563a721fb62a31366b66"},
}

func GetAdInfo(adId string) *ADInfo {
	for _, v := range g_adInfos {
		if v.AdId == adId {
			return v
		}
	}
	return nil
}

func GetAllAdInfo() []*ADInfo {
	return g_adInfos
}

type ADData struct {
	TS     int64
	AdId   string
	Show   int32
	Click  int32
	Income float64
}

func HandleThirdPartyData(tp int32, filename string) error {
	if tp < AD_CSJ && tp > AD_XM {
		return errors.New("Third party not support!")
	}
	adDatas, err := parseAdData(tp, filename)
	if err != nil {
		return err
	}
	if len(adDatas) > 0 {
		return writeAdDatas(adDatas)
	}
	return nil
}

func GetAdDatas(start, end int64, adIds []string) []*ADData {
	adDatas := make([]*ADData, 0, 1024)
	if start > end || len(adIds) == 0 {
		return adDatas
	}
	file, err := os.OpenFile("./upload/third_party.dat", os.O_RDONLY, 0644)
	if err != nil {
		logs.Error("[GetAdDatas] Open file error:", err)
		return adDatas
	}
	defer file.Close()
	buf := bufio.NewReader(file)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				logs.Error("[GetAdDatas] ReadString error:", err)
				return adDatas
			}
		}
		ad := &ADData{}
		sl := strings.Split(line[:len(line)-1], "|")
		if len(sl) < 5 {
			logs.Error("[GetAdDatas] data file broken!")
			return adDatas
		}
		ad.TS, _ = strconv.ParseInt(sl[0], 10, 32)
		ad.AdId = sl[1]
		show, _ := strconv.ParseInt(sl[2], 10, 32)
		ad.Show = int32(show)
		click, _ := strconv.ParseInt(sl[3], 10, 32)
		ad.Click = int32(click)
		income, _ := strconv.ParseFloat(sl[4], 32)
		ad.Income = income
		if ad.TS >= start && ad.TS <= end {
			for _, v := range adIds {
				if v == ad.AdId {
					adDatas = append(adDatas, ad)
					break
				}
			}
		}
	}
	return sortAndUniqueAdDatas(adDatas)
}

func sortAndUniqueAdDatas(adDatas []*ADData) []*ADData {
	if len(adDatas) < 2 {
		return adDatas
	}
	sort.Slice(adDatas, func(i, j int) bool {
		if adDatas[i].TS < adDatas[j].TS {
			return true
		} else if adDatas[i].TS > adDatas[j].TS {
			return false
		} else {
			return adDatas[i].AdId < adDatas[j].AdId
		}
	})
	for i := 1; i < len(adDatas); {
		if adDatas[i-1].TS == adDatas[i].TS && adDatas[i-1].AdId == adDatas[i].AdId {
			adDatas = append(adDatas[:i], adDatas[i+1:]...)
		} else {
			i++
		}
	}
	return adDatas
}

func writeAdDatas(adDatas []*ADData) error {
	g_tpLock.Lock()
	f, err := os.OpenFile("./data/third_party.dat", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		g_tpLock.Unlock()
		return err
	}
	defer f.Close()
	for _, v := range adDatas {
		content := fmt.Sprintf("%d|%s|%d|%d|%f\n", v.TS, v.AdId, v.Show, v.Click, v.Income)
		_, err = f.Write([]byte(content))
		if err != nil {
			g_tpLock.Unlock()
			return err
		}
	}
	g_tpLock.Unlock()
	return nil
}

func parseAdData(tp int32, filename string) ([]*ADData, error) {
	xlFile, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, err
	}
	sheetName := xlFile.GetSheetName(0)
	if sheetName == "" {
		return nil, errors.New("Empty sheet!")
	}
	rows := xlFile.GetRows(sheetName)
	if len(rows) == 0 {
		return nil, errors.New("Empty row!")
	}
	head := rows[0]
	iTs, iAdId, iShow, iClick, iIncome := -1, -1, -1, -1, -1
	switch tp {
	case AD_CSJ:
		for i, cell := range head {
			switch cell {
			case "时间":
				iTs = i
			case "代码位ID":
				iAdId = i
			case "展现量":
				iShow = i
			case "点击量":
				iClick = i
			case "预估收益(人民币)":
				iIncome = i
			}
		}
	case AD_YLH:
		for i, cell := range head {
			switch cell {
			case "时间":
				iTs = i
			case "广告位ID":
				iAdId = i
			case "广告展示数":
				iShow = i
			case "点击量":
				iClick = i
			case "预计收入":
				iIncome = i
			}
		}
	case AD_XM:
		for i, cell := range head {
			switch cell {
			case "日期":
				iTs = i
			case "广告位ID":
				iAdId = i
			case "展现量":
				iShow = i
			case "点击数":
				iClick = i
			case "预计收入(元)":
				iIncome = i
			}
		}
	}
	if iTs == -1 || iAdId == -1 || iShow == -1 || iClick == -1 || iIncome == -1 {
		return nil, errors.New("Data invalid!")
	}
	adDatas := []*ADData{}
	start := 1
	if tp == AD_CSJ {
		start = 2
	}
	for i := start; i < len(rows); i++ {
		ad := &ADData{}
		if iTs < len(rows[i]) {
			date := rows[i][iTs]
			if tp == AD_YLH {
				ss := strings.Split(date, "-")
				if len(ss) < 3 {
					return nil, errors.New("YLH date invalid!")
				}
				date = "20" + ss[2] + "-" + ss[0] + "-" + ss[1]
			} else if tp == AD_XM {
				date = date[:4] + "-" + date[4:6] + "-" + date[6:]
			}
			ad.TS = util.Date2ts(date)
		}
		if iAdId < len(rows[i]) {
			ad.AdId = util.RemoveSideBlank(rows[i][iAdId])
		}
		if iShow < len(rows[i]) {
			n, err := strconv.ParseInt(rows[i][iShow], 10, 32)
			if err != nil {
				return nil, err
			}
			ad.Show = int32(n)
		}
		if iClick < len(rows[i]) {
			n, err := strconv.ParseInt(rows[i][iClick], 10, 32)
			if err != nil {
				return nil, err
			}
			ad.Click = int32(n)
		}
		if iIncome < len(rows[i]) {
			n, err := strconv.ParseFloat(rows[i][iIncome], 64)
			if err != nil {
				return nil, err
			}
			ad.Income = n
		}
		adDatas = append(adDatas, ad)
	}
	return adDatas, nil
}

var g_tpLock sync.Mutex
