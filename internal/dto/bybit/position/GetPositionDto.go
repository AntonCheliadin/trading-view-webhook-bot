package position

type GetPositionDto struct {
	RetCode int    `mapstructure:"retCode"`
	RetMsg  string `mapstructure:"retMsg"`
	Result  struct {
		Category       string        `mapstructure:"category"`
		List           []PositionDto `mapstructure:"list"`
		NextPageCursor string        `mapstructure:"nextPageCursor"`
	} `mapstructure:"result"`
	RetExtInfo struct {
	} `mapstructure:"retExtInfo"`
	Time int64 `mapstructure:"time"`
}

type PositionDto struct {
	AdlRankIndicator       int    `mapstructure:"adlRankIndicator"`
	AutoAddMargin          int    `mapstructure:"autoAddMargin"`
	AvgPrice               string `mapstructure:"avgPrice"`
	BustPrice              string `mapstructure:"bustPrice"`
	CreatedTime            string `mapstructure:"createdTime"`
	CumRealisedPnl         string `mapstructure:"cumRealisedPnl"`
	CurRealisedPnl         string `mapstructure:"curRealisedPnl"`
	IsReduceOnly           bool   `mapstructure:"isReduceOnly"`
	Leverage               string `mapstructure:"leverage"`
	LeverageSysUpdatedTime string `mapstructure:"leverageSysUpdatedTime"`
	LiqPrice               string `mapstructure:"liqPrice"`
	MarkPrice              string `mapstructure:"markPrice"`
	MmrSysUpdatedTime      string `mapstructure:"mmrSysUpdatedTime"`
	PositionBalance        string `mapstructure:"positionBalance"`
	PositionIM             string `mapstructure:"positionIM"`
	PositionIdx            int    `mapstructure:"positionIdx"`
	PositionMM             string `mapstructure:"positionMM"`
	PositionStatus         string `mapstructure:"positionStatus"`
	PositionValue          string `mapstructure:"positionValue"`
	RiskId                 int    `mapstructure:"riskId"`
	RiskLimitValue         string `mapstructure:"riskLimitValue"`
	Seq                    int64  `mapstructure:"seq"`
	SessionAvgPrice        string `mapstructure:"sessionAvgPrice"`
	Side                   string `mapstructure:"side"`
	Size                   string `mapstructure:"size"`
	StopLoss               string `mapstructure:"stopLoss"`
	Symbol                 string `mapstructure:"symbol"`
	TakeProfit             string `mapstructure:"takeProfit"`
	TpslMode               string `mapstructure:"tpslMode"`
	TradeMode              int    `mapstructure:"tradeMode"`
	TrailingStop           string `mapstructure:"trailingStop"`
	UnrealisedPnl          string `mapstructure:"unrealisedPnl"`
	UpdatedTime            string `mapstructure:"updatedTime"`
}
