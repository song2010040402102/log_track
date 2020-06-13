package stat

import "fmt"

// WheelOutputHandler 大转盘次数产出
func WheelOutputHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	title := fmt.Sprintf("大转盘产出(%s ~ %s)", startDate, endDate)
	return currencyHandler(gameId, LOG_RES_OUTPUT, startDate, endDate, channel, device, "WheelFlow", title)
}

// WheelConsumeHandler 大转盘次数消耗
func WheelConsumeHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	title := fmt.Sprintf("大转盘消耗(%s ~ %s)", startDate, endDate)
	return currencyHandler(gameId, LOG_RES_CONSUME, startDate, endDate, channel, device, "WheelFlow", title)
}
