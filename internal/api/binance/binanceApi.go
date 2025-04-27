package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
	"tradingViewWebhookBot/internal/api"
	"tradingViewWebhookBot/internal/constants/futureType"
	"tradingViewWebhookBot/internal/domain"
	"tradingViewWebhookBot/internal/dto/binance"
	"tradingViewWebhookBot/internal/util"
)

func NewBinanceApi() api.ExchangeApi {
	return &BinanceApi{}
}

// https://binance-docs.github.io/apidocs/spot/en/#test-connectivity
type BinanceApi struct {
}

func (api *BinanceApi) GetKlines(coin *domain.Coin, interval string, limit int, fromTime time.Time) (api.KlinesDto, error) {
	return nil, errors.New("Not implemented for Binance API")
}

func (api *BinanceApi) GetKlinesFutures(coin *domain.Coin, interval string, limit int, fromTime time.Time) (api.KlinesDto, error) {
	return nil, errors.New("Not implemented for Binance API")
}

func (api *BinanceApi) GetCurrentCoinPriceForFutures(coin *domain.Coin) (float64, error) {
	return 0, errors.New("Not implemennted.")
}

func (api *BinanceApi) GetCurrentCoinPrice(coin *domain.Coin) (float64, error) {
	resp, err := http.Get("https://api.binance.com/api/v3/ticker/price?symbol=" + coin.Symbol)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var priceDto binance.PriceDto
	if err := json.NewDecoder(resp.Body).Decode(&priceDto); err != nil {
		return 0, err
	}

	return priceDto.GetPrice()
}

func (api *BinanceApi) BuyCoinByMarket(coin *domain.Coin, amount float64, price float64) (api.OrderResponseDto, error) {
	queryParams := api.buildParams(coin, amount, "BUY")
	return api.orderCoinByMarket(queryParams)
}

func (api *BinanceApi) SellCoinByMarket(coin *domain.Coin, amount float64, price float64) (api.OrderResponseDto, error) {
	queryParams := api.buildParams(coin, amount, "SELL")
	return api.orderCoinByMarket(queryParams)
}

func (api *BinanceApi) orderCoinByMarket(queryParams string) (api.OrderResponseDto, error) {
	zap.S().Debugf("OrderCoinByMarket = %s", queryParams)

	uri := "https://api.binance.com/api/v3/order?" // /test
	signatureParameter := "&signature=" + api.sign(queryParams)

	url := uri + queryParams + signatureParameter

	method := "POST"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		zap.S().Errorf("API error: %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-MBX-APIKEY", os.Getenv("BINANCE_API_KEY"))

	res, err := client.Do(req)
	if err != nil {
		zap.S().Errorf("API error: %s", err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		zap.S().Errorf("API error: %s", err)
		return nil, err
	}
	zap.S().Debugf("API response: %s", string(body))

	dto := binance.OrderResponseBinanceDto{}
	errUnmarshal := json.Unmarshal(body, &dto)
	if errUnmarshal != nil {
		zap.S().Error("Unmarshal error", errUnmarshal.Error())
		return nil, errUnmarshal
	}

	return dto, nil
}

func (api *BinanceApi) buildParams(coin *domain.Coin, amount float64, side string) string {
	return "symbol=" + coin.Symbol +
		"&side=" + side +
		"&type=MARKET" +
		"&recvWindow=60000" +
		"&quantity=" + strings.TrimRight(fmt.Sprintf("%f", amount), "0") +
		"&timestamp=" + util.MakeTimestamp()
}

func (api *BinanceApi) sign(data string) string {
	secret := os.Getenv("BINANCE_SECRET_KEY")

	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(secret))

	// Write Data to it
	h.Write([]byte(data))

	// Get result and encode as hexadecimal string
	sha := hex.EncodeToString(h.Sum(nil))

	return sha
}

func (api *BinanceApi) OpenFuturesOrder(coin *domain.Coin, amount float64, price float64, futuresType futureType.FuturesType, stopLossPriceInCents float64) (api.OrderResponseDto, error) {
	return nil, errors.New("Futures api is not implemented")
}
func (api *BinanceApi) CloseFuturesOrder(coin *domain.Coin, openedTransaction *domain.Transaction, price float64) (api.OrderResponseDto, error) {
	return nil, errors.New("Futures api is not implemented")
}

func (api *BinanceApi) GetWalletBalance() (api.WalletBalanceDto, error) {
	return nil, nil
}

func (api *BinanceApi) SetFuturesLeverage(coin *domain.Coin, leverage int) error {
	return nil
}

func (api *BinanceApi) SetIsolatedMargin(coin *domain.Coin, leverage int) error {
	return nil
}

func (api *BinanceApi) IsFuturesPositionOpened(coin *domain.Coin, openedOrder *domain.Transaction) bool {
	return true
}
func (api *BinanceApi) GetCloseTradeRecord(coin *domain.Coin, openTransaction *domain.Transaction) (api.OrderResponseDto, error) {
	return nil, nil
}

func (api *BinanceApi) GetLastFuturesOrder(coin *domain.Coin, clientOrderId string) (api.OrderResponseDto, error) {
	return nil, nil
}

func (api *BinanceApi) SetApiKey(apiKey string) {
}

func (api *BinanceApi) SetSecretKey(secretKey string) {
}
