package stat

import "fmt"

// StarIntegalOutputHandler 段位分次数产出
func StarIntegalOutputHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	title := fmt.Sprintf("段位分产出(%s ~ %s)", startDate, endDate)
	return currencyHandler(gameId, LOG_RES_OUTPUT, startDate, endDate, channel, device, "StarIntegalFlow", title)
}

// StarIntegalConsumeHandler 段位分次数消耗
func StarIntegalConsumeHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	title := fmt.Sprintf("段位分消耗(%s ~ %s)", startDate, endDate)
	return currencyHandler(gameId, LOG_RES_CONSUME, startDate, endDate, channel, device, "StarIntegalFlow", title)
}
