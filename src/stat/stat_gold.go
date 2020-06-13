package stat

import "fmt"

// GoldOutputHandler 金币产出
func GoldOutputHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	title := fmt.Sprintf("金币产出(%s ~ %s)", startDate, endDate)
	return currencyHandler(gameId, LOG_RES_OUTPUT, startDate, endDate, channel, device, "GoldFlow", title)
}

// GoldConsumeHandler 金币消耗
func GoldConsumeHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	title := fmt.Sprintf("金币消耗(%s ~ %s)", startDate, endDate)
	return currencyHandler(gameId, LOG_RES_CONSUME, startDate, endDate, channel, device, "GoldFlow", title)
}
