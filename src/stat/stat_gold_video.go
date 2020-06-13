package stat

import (
	"common"
	"data"
	"fmt"
	"github.com/astaxie/beego/logs"
	"math"
	"strconv"
	"util"
)

type goldRoomRank struct {
	title string
	low   int32
	high  int32
}

type playerWatchInfo struct {
	watchTimes int32
	playTimes  int32
	gold       int32
}

type goldRoomVideoStat struct {
	watchTimes  int32
	watchPerson map[string]bool
	playTimes   int32
	playPerson  map[string]bool
}

var goldRoomVideoType = []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 22, 24, 29, 30}

func GoldRoomVideoHandler(gameId int32, startDate, endDate, channel string, device string) *ItemResult {
	var roomRanks = []*goldRoomRank{
		&goldRoomRank{title: "1k ~ 3k", low: 1000, high: 2999},
		&goldRoomRank{title: "3k ~ 8k", low: 3000, high: 7999},
		&goldRoomRank{title: "8k ~ 100k", low: 8000, high: 99999},
		&goldRoomRank{title: "100k ~ 1m", low: 100000, high: 999999},
		&goldRoomRank{title: "1m ~ ", low: 1000000, high: math.MaxInt32},
		&goldRoomRank{title: "汇总", low: 0, high: 0},
	}

	var playerWatch map[string]*playerWatchInfo
	var watchs []map[string]*playerWatchInfo
	err := data.GetLogData(gameId, "GoldFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos == 0 {
			playerWatch = make(map[string]*playerWatchInfo)
		} else if pos > 0 {
			if len(s) < 18 {
				logs.Error("GoldFlow length error", len(s))
				return false
			}
			if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
				return true
			}
			source, _ := strconv.Atoi(s[14])
			gold, _ := strconv.Atoi(s[11])

			// 金币场游玩
			if source == 7 {
				if _, ok := playerWatch[s[0]]; !ok {
					playerWatch[s[0]] = &playerWatchInfo{
						playTimes:  0,
						watchTimes: 0,
						gold:       0,
					}
				}
				playerWatch[s[0]].playTimes++
				playerWatch[s[0]].gold = int32(gold)
			}
		} else {
			watchs = append(watchs, playerWatch)
		}
		return true
	})
	if err != nil {
		return nil
	}
	err = data.GetLogData(gameId, "VideoStartFlow", startDate, endDate, func(day int32, pos int32, s []string) bool {
		if pos == 0 {
			playerWatch = watchs[day]
		} else if pos > 0 {
			if len(s) < 11 {
				logs.Error("video start data error!")
				return false
			}
			if !common.IsMultiCond(channel, s[8]) || !common.IsMultiCond(device, s[9]) {
				return true
			}
			videoType, _ := strconv.Atoi(s[10])
			if isgoldRoomVideoType(int32(videoType)) {
				if player, ok := playerWatch[s[0]]; ok {
					player.watchTimes++
				} else {
					logs.Error("player did not included", s[0], "videoType", videoType)
				}
			}
		}
		return true
	})
	if err != nil {
		return nil
	}

	var stats []map[string]*goldRoomVideoStat
	for _, watch := range watchs {
		st := make(map[string]*goldRoomVideoStat)
		for _, rank := range roomRanks {
			st[rank.title] = &goldRoomVideoStat{
				watchTimes:  0,
				watchPerson: make(map[string]bool),
				playTimes:   0,
				playPerson:  make(map[string]bool),
			}
		}
		sumTitle := roomRanks[len(roomRanks)-1].title
		for name, info := range watch {
			for _, rank := range roomRanks {
				if info.gold < rank.high && info.gold >= rank.low {
					if info.playTimes > 0 {
						st[rank.title].playTimes += info.playTimes
						st[rank.title].playPerson[name] = true
						st[sumTitle].playTimes += info.playTimes
						st[sumTitle].playPerson[name] = true
					}
					if info.watchTimes > 0 {
						st[rank.title].watchTimes += info.watchTimes
						st[rank.title].watchPerson[name] = true
						st[sumTitle].watchTimes += info.watchTimes
						st[sumTitle].watchPerson[name] = true
					}
				}
			}
		}
		stats = append(stats, st)
	}

	heads := []string{"日期", "渠道", "设备", "金币档位", "视频观看次数", "视频观看人数", "游玩次数",
		"游玩人数", "平均游玩次数", "观看视频率", "所有玩家人均观看数", "看视频玩家人均观看数"}
	var items [][]string
	for i, st := range stats {
		date := util.Ts2date(util.Date2ts(startDate) + int64(i*86400))
		for j := 0; j < len(roomRanks); j++ {
			rankData := st[roomRanks[j].title]
			var row []string
			if j == 0 {
				row = append(row, date)
			} else {
				row = append(row, "")
			}
			var avgWatchTimes, avgPlayTimes, ratioWatchPlay, avgWatchTimesAll float64
			if len(rankData.watchPerson) > 0 {
				avgWatchTimes = float64(rankData.watchTimes) / float64(len(rankData.watchPerson))
			}
			if len(rankData.playPerson) > 0 {
				avgPlayTimes = float64(rankData.playTimes) / float64(len(rankData.playPerson))
				ratioWatchPlay = float64(len(rankData.watchPerson)) / float64(len(rankData.playPerson))
				avgWatchTimesAll = float64(rankData.watchTimes) / float64(len(rankData.playPerson))
			}
			data := []string{
				channel,
				device,
				roomRanks[j].title,
				strconv.FormatInt(int64(rankData.watchTimes), 10),
				strconv.FormatInt(int64(len(rankData.watchPerson)), 10),
				strconv.FormatInt(int64(rankData.playTimes), 10),
				strconv.FormatInt(int64(len(rankData.playPerson)), 10),
				strconv.FormatFloat(avgPlayTimes, 'f', 2, 64),
				fmt.Sprintf("%.2f%%", ratioWatchPlay*100),
				strconv.FormatFloat(avgWatchTimesAll, 'f', 2, 64),
				strconv.FormatFloat(avgWatchTimes, 'f', 2, 64),
			}
			row = append(row, data...)
			items = append(items, row)
		}
	}
	return &ItemResult{
		TTable: NewTable(fmt.Sprintf("金币场玩家观看视频统计(%s ~ %s)", startDate, endDate), heads, items),
	}
}

func isgoldRoomVideoType(channel int32) bool {
	for _, v := range goldRoomVideoType {
		if v == channel {
			return true
		}
	}
	return false
}
