package bybit

import (
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"time"
	"tradingViewWebhookBot/internal/api"
	"tradingViewWebhookBot/internal/util"
)

type KlinesDto struct {
	RetCode int        `json:"ret_code"`
	RetMsg  string     `json:"ret_msg"`
	ExtCode string     `json:"ext_code"`
	ExtInfo string     `json:"ext_info"`
	Result  []KlineDto `json:"result"`
	TimeNow string     `json:"time_now"`
}

func (d *KlinesDto) String() string {
	return fmt.Sprintf("KlinesDto {RetCode: %v, RetMsg: %v, ExtCode: %v, ExtInfo: %v, TimeNow: %v, ResultSize: %v}",
		d.RetCode, d.RetMsg, d.ExtCode, d.ExtInfo, d.TimeNow, len(d.Result))
}

func (dto *KlinesDto) GetKlines() []api.KlineDto {
	castedKlines := make([]api.KlineDto, len(dto.Result), len(dto.Result))
	for i := range dto.Result {
		castedKlines[i] = &dto.Result[i]
	}

	return castedKlines
}

type KlineDto struct {
	Id       int     `json:"id"`
	Symbol   string  `json:"symbol"`
	Period   string  `json:"period"`
	StartAt  int     `json:"start_at"` // Start timestamp point for result, in seconds
	Volume   float64 `json:"volume"`
	Open     float64 `json:"open"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Close    float64 `json:"close"`
	Interval string  `json:"interval"`
	OpenTime int     `json:"open_time"`
	Turnover float64 `json:"turnover"`
}

func (dto *KlineDto) GetSymbol() string {
	return dto.Symbol
}

func (dto *KlineDto) GetInterval() string {
	return dto.Interval
}

func (dto *KlineDto) GetStartAt() time.Time {
	return util.GetTimeBySeconds(dto.StartAt)
}

func (dto *KlineDto) GetCloseAt() time.Time {
	parseInt, err := strconv.ParseInt(dto.Interval, 10, 64)
	if err != nil {
		zap.S().Errorf("Error: %s", err.Error())
	}
	return dto.GetStartAt().Add(time.Minute * time.Duration(parseInt))
}

func (dto *KlineDto) GetOpen() float64 {
	return dto.Open
}

func (dto *KlineDto) GetHigh() float64 {
	return dto.High
}

func (dto *KlineDto) GetLow() float64 {
	return dto.Low
}
func (dto *KlineDto) GetClose() float64 {
	return dto.Close
}
