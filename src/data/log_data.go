package data

import (
	"bufio"
	"common"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"util"
)

var ALL_FLOW []string = []string{"WxCreate", "Create", "Login", "Logout", "Online", "AdToMinGameFlow", "AdToMinGameResultFlow",
	"DiamondFlow", "GoldFlow", "FangkaFlow", "VoucherFlow", "EnrollVoucherFlow", "StarIntegalFlow",
	"ZjjIntegalFlow", "TableStatFlow", "VipExpFlow", "WheelFlow", "MaterialFlow", "TaskFlow", "TaskFinishFlow", "TaskDrawFlow",
	"PayFlow", "RedbagConsumeFlow", "IncomeFlow", "IncomeTelebillFlow", "RedBagFlow", "TeleBillFlow",
	"DrawCashFlow", "RoomFlow", "RealRoomFlow", "RoomNoBenifitFlow", "RoomMergePlayFlow", "RoomAutoFlow",
	"MatchFlow", "MatchResultFlow", "GrandPrixFlow", "GrandPrixResultFlow", "WeChatSharePicFlow",
	"VideoStartFlow", "VideoEndFlow", "VideoClickFlow", "VideoLoginFlow", "VideoLoginClickFlow",
	"VideoInsertFlow", "VideoInsertClickFlow", "BannerStartFlow", "BannerEndFlow", "BannerClickFlow",
	"WeChatShareClickFlow", "WeChatShareLoginFlow",
}

func IS_VALID_FLOW(flow string) bool {
	for _, v := range ALL_FLOW {
		if v == flow {
			return true
		}
	}
	return false
}

const SERVER_IP string = "122.226.109.132" //数据源地址
const LOG_SEP string = "|"                 //日志字段分隔符

func GetLogData(gameId int32, flow string, startDate, endDate string, callback func(int32, int32, []string) bool) error {
	if callback == nil {
		return errors.New("callback cannot nil!")
	}
	if !IS_VALID_FLOW(flow) {
		return errors.New("invalid flow!")
	}
	today := util.GetDate()
	if startDate == "" {
		startDate = today
	}
	if endDate == "" || util.Date2ts(endDate) > time.Now().Unix() {
		endDate = today
	}
	days := (util.Date2ts(endDate)-util.Date2ts(startDate))/86400 + 1
	if days <= 0 {
		return errors.New("invalid date!")
	}
	g_dateLock.Lock()
	serv_dates := []string{} //需要从服务器获取数据的日期
	for i := int64(0); i < days; i++ {
		if i == days-1 && endDate == today { //今天的数据实时在变，不能直接用缓存
			serv_dates = append(serv_dates, today)
		} else {
			date := util.Ts2date(util.Date2ts(startDate) + i*86400)
			if !util.IsFileExist(makeFileName(gameId, flow, date)) { //过去的文件不存在，则从服务器获取
				serv_dates = append(serv_dates, date)
			}
		}
	}
	if len(serv_dates) > 0 {
		getServerData(gameId, flow, serv_dates)
	}
	g_dateLock.Unlock()

	for i := int64(0); i < days; i++ {
		filename := makeFileName(gameId, flow, util.Ts2date(util.Date2ts(startDate)+i*86400))
		if err := readFromFile(filename, int32(i), callback); err != nil {
			return err
		}
	}
	return nil
}

func AutoGetLogData() {
	logs.Notice("auto get data")
	yesterday := util.Ts2date(time.Now().Unix() - 86400)
	for _, gameId := range common.GetAllGameId() {
		for _, flow := range ALL_FLOW {
			g_dateLock.Lock()
			getServerData(gameId, flow, []string{yesterday})
			g_dateLock.Unlock()
		}
	}
}

func getServerData(gameId int32, flow string, serv_dates []string) {
	var wg sync.WaitGroup
	for _, v := range serv_dates {
		date := v
		wg.Add(1)
		go func() {
			defer wg.Done()
			url := makeUrl(gameId, flow, date, date)
			if url == "" {
				return
			}
			resp, err := http.Get(url)
			if err != nil {
				logs.Error("[getServerData] http get error:", err, "url: ", url)
				return
			}
			defer resp.Body.Close()
			util.SaveToFile(resp.Body, makeFileName(gameId, flow, date))
		}()
	}
	wg.Wait()
}

func readFromFile(filename string, day int32, callback func(int32, int32, []string) bool) error {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		logs.Error("[readFromFile] OpenFile error:", err)
		return err
	}
	defer f.Close()

	pos := int32(0)
	callback(day, 0, nil)
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				logs.Error("[readFromFile] ReadString error:", err)
				return err
			}
		}
		pos++
		if !callback(day, pos, strings.Split(line[:len(line)-1], LOG_SEP)) {
			return errors.New("callback break!")
		}
	}
	callback(day, -1, nil)
	return nil
}

func makeUrl(gameId int32, flow string, startDate, endDate string) string {
	return fmt.Sprintf("http://%s/rest/data/%s?server_id=%s&game_id=%s&start_time=%s&end_time=%s",
		SERVER_IP, flow, common.GetAllServerId()[gameId], common.GetAllServerName()[gameId], startDate, endDate)
}

func makeFileName(gameId int32, flow string, date string) string {
	return fmt.Sprintf("data/log/%d_%s_%s.dat", gameId, flow, date)
}

var g_dateLock sync.Mutex
