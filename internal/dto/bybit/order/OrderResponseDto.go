package order

type OrderResponseDto struct {
	RetCode    int         `mapstructure:"retCode"`
	RetMsg     string      `mapstructure:"retMsg"`
	Result     OrderResult `mapstructure:"result"`
	RetExtInfo interface{} `mapstructure:"retExtInfo"`
	Time       int64       `mapstructure:"time"`
}

type OrderResult struct {
	OrderId     string `mapstructure:"orderId"`
	OrderLinkId string `mapstructure:"orderLinkId"`
}
