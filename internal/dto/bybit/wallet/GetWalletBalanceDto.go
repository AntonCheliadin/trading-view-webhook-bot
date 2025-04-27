package wallet

import "strconv"

type GetWalletBalanceDto struct {
	RetCode int    `mapstructure:"retCode"`
	RetMsg  string `mapstructure:"retMsg"`
	Result  struct {
		List []struct {
			TotalEquity            string `mapstructure:"totalEquity"`
			AccountIMRate          string `mapstructure:"accountIMRate"`
			TotalMarginBalance     string `mapstructure:"totalMarginBalance"`
			TotalInitialMargin     string `mapstructure:"totalInitialMargin"`
			AccountType            string `mapstructure:"accountType"`
			TotalAvailableBalance  string `mapstructure:"totalAvailableBalance"`
			AccountMMRate          string `mapstructure:"accountMMRate"`
			TotalPerpUPL           string `mapstructure:"totalPerpUPL"`
			TotalWalletBalance     string `mapstructure:"totalWalletBalance"`
			AccountLTV             string `mapstructure:"accountLTV"`
			TotalMaintenanceMargin string `mapstructure:"totalMaintenanceMargin"`
			Coin                   []struct {
				AvailableToBorrow   string `mapstructure:"availableToBorrow"`
				Bonus               string `mapstructure:"bonus"`
				AccruedInterest     string `mapstructure:"accruedInterest"`
				AvailableToWithdraw string `mapstructure:"availableToWithdraw"`
				TotalOrderIM        string `mapstructure:"totalOrderIM"`
				Equity              string `mapstructure:"equity"`
				TotalPositionMM     string `mapstructure:"totalPositionMM"`
				UsdValue            string `mapstructure:"usdValue"`
				SpotHedgingQty      string `mapstructure:"spotHedgingQty"`
				UnrealisedPnl       string `mapstructure:"unrealisedPnl"`
				CollateralSwitch    bool   `mapstructure:"collateralSwitch"`
				BorrowAmount        string `mapstructure:"borrowAmount"`
				TotalPositionIM     string `mapstructure:"totalPositionIM"`
				WalletBalance       string `mapstructure:"walletBalance"`
				CumRealisedPnl      string `mapstructure:"cumRealisedPnl"`
				Locked              string `mapstructure:"locked"`
				MarginCollateral    bool   `mapstructure:"marginCollateral"`
				Coin                string `mapstructure:"coin"`
			} `mapstructure:"coin"`
		} `mapstructure:"list"`
	} `mapstructure:"result"`
	RetExtInfo struct {
	} `mapstructure:"retExtInfo"`
	Time int64 `mapstructure:"time"`
}

func (dto *GetWalletBalanceDto) GetAvailableBalance() float64 {
	return dto.parseWalletBalance() - dto.parseTotalPositionIM() - dto.parseTotalOrderIM() - dto.parseLocked()
}

func (dto *GetWalletBalanceDto) parseWalletBalance() float64 {
	if val, err := strconv.ParseFloat(dto.Result.List[0].Coin[0].UsdValue, 64); err == nil {
		return val
	}
	return 0
}

func (dto *GetWalletBalanceDto) parseTotalPositionIM() float64 {
	if val, err := strconv.ParseFloat(dto.Result.List[0].Coin[0].TotalPositionIM, 64); err == nil {
		return val
	}
	return 0
}

func (dto *GetWalletBalanceDto) parseTotalOrderIM() float64 {
	if val, err := strconv.ParseFloat(dto.Result.List[0].Coin[0].TotalOrderIM, 64); err == nil {
		return val
	}
	return 0
}

func (dto *GetWalletBalanceDto) parseLocked() float64 {
	if val, err := strconv.ParseFloat(dto.Result.List[0].Coin[0].Locked, 64); err == nil {
		return val
	}
	return 0
}
