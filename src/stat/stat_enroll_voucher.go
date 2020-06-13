package stat

import "fmt"

// EnrollVoucherOutputHandler 报名券产出
func EnrollVoucherOutputHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	title := fmt.Sprintf("报名券产出(%s ~ %s)", startDate, endDate)
	return currencyHandler(gameId, LOG_RES_OUTPUT, startDate, endDate, channel, device, "EnrollVoucherFlow", title)
}

// EnrollVoucherConsumeHandler 报名券消耗
func EnrollVoucherConsumeHandler(gameId int32, startDate, endDate, channel, device string) *ItemResult {
	title := fmt.Sprintf("报名券消耗(%s ~ %s)", startDate, endDate)
	return currencyHandler(gameId, LOG_RES_CONSUME, startDate, endDate, channel, device, "EnrollVoucherFlow", title)
}
