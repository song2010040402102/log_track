package stat

import (
	"fmt"
)

// DiamondOutputHandler 钻石产出
func DiamondOutputHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	title := fmt.Sprintf("钻石产出(%s ~ %s)", startDate, endDate)
	return currencyHandler(gameId, LOG_RES_OUTPUT, startDate, endDate, channel, device, "DiamondFlow", title)
}

// DiamondConsumeHandler 钻石消耗
func DiamondConsumeHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	title := fmt.Sprintf("钻石消耗(%s ~ %s)", startDate, endDate)
	return currencyHandler(gameId, LOG_RES_CONSUME, startDate, endDate, channel, device, "DiamondFlow", title)
}
