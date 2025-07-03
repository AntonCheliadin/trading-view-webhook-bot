package bybit

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	bybit "github.com/bybit-exchange/bybit.go.api"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"tradingViewWebhookBot/internal/api"
	"tradingViewWebhookBot/internal/constants/futureType"
	"tradingViewWebhookBot/internal/domain"
	bybitDto "tradingViewWebhookBot/internal/dto/bybit"
	"tradingViewWebhookBot/internal/dto/bybit/order"
	"tradingViewWebhookBot/internal/dto/bybit/position"
	"tradingViewWebhookBot/internal/dto/bybit/wallet"
	"tradingViewWebhookBot/internal/util"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

func NewBybitApi(apiKey string, secretKey string) api.ExchangeApi {
	return &BybitApi{
		apiKey:    apiKey,
		secretKey: secretKey,
		client:    bybit.NewBybitHttpClient(apiKey, secretKey, bybit.WithBaseURL(bybit.MAINNET)),
	}
}

type BybitApi struct {
	apiKey    string
	secretKey string
	client    *bybit.Client
}

func (bybitApi *BybitApi) GetKlines(coin *domain.Coin, interval string, limit int, fromTime time.Time) (api.KlinesDto, error) {
	resp, err := http.Get("https://api.bytick.com/public/linear/kline?" +
		"symbol=" + coin.Symbol +
		"&interval=" + interval +
		"&limit=" + strconv.Itoa(limit) +
		"&from=" + strconv.Itoa(util.GetSecondsByTime(fromTime)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var dto bybitDto.KlinesDto
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, err
	}

	return &dto, nil
}
func (bybitApi *BybitApi) GetKlinesFutures(coin *domain.Coin, interval string, limit int, fromTime time.Time) (api.KlinesDto, error) {
	intervalInt, _ := strconv.Atoi(interval)
	end := fromTime.Add(time.Minute * time.Duration(intervalInt*limit))

	resp, err := http.Get("https://api.bytick.com/derivatives/v3/public/kline?" +
		"category=linear" +
		"&symbol=" + coin.Symbol +
		"&interval=" + interval +
		"&start=" + strconv.FormatInt(util.GetMillisByTime(fromTime), 10) +
		"&end=" + strconv.FormatInt(util.GetMillisByTime(end), 10) +
		"&limit=" + strconv.Itoa(limit))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dto := &bybitDto.KlinesFuturesDto{Interval: intervalInt}
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, err
	}

	return dto, nil
}

func (api *BybitApi) GetCurrentCoinPriceForFutures(coin *domain.Coin) (float64, error) {
	resp, err := http.Get("https://api.bytick.com/derivatives/v3/public/tickers?symbol=" + coin.Symbol)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	var priceDto bybitDto.TickerInfoDto
	if err := json.NewDecoder(resp.Body).Decode(&priceDto); err != nil {
		return 0, err
	}

	return priceDto.Price()
}

func (api *BybitApi) GetCurrentCoinPrice(coin *domain.Coin) (float64, error) {
	params := map[string]interface{}{
		"category": "linear", // Important: "linear" = USDT perpetual
		"symbol":   coin.Symbol,
		"interval": "1",
		"limit":    "1",
	}

	priceResult, err := api.client.NewUtaBybitServiceWithParams(params).GetMarkPriceKline(context.Background())
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	// convert the interface{} to a map
	resultMap, ok := priceResult.Result.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("failed to convert result to map")
	}

	// Get the list array
	list, ok := resultMap["list"].([]interface{})
	if !ok {
		return 0, fmt.Errorf("failed to get list from result")
	}

	// Get the first item from list (which is itself an array)
	firstItem, ok := list[0].([]interface{})
	if !ok {
		return 0, fmt.Errorf("failed to get first item from list")
	}

	// Get the last value (index 4) and convert to string
	lastValue, ok := firstItem[4].(string)
	if !ok {
		return 0, fmt.Errorf("failed to get last value as string")
	}

	// Convert string to float64
	price, err := strconv.ParseFloat(lastValue, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price: %v", err)
	}

	return price, nil
}

func (api *BybitApi) BuyCoinByMarket(coin *domain.Coin, quantity float64, price float64) (api.OrderResponseDto, error) {
	return nil, errors.New("Not implemennted.")

}

func (api *BybitApi) SellCoinByMarket(coin *domain.Coin, quantity float64, price float64) (api.OrderResponseDto, error) {
	return nil, errors.New("Not implemennted.")
}

func (api *BybitApi) getOrderById(orderId string) (*order.OrderHistoryDto, error) {
	params := map[string]interface{}{
		"category": "linear",
		"orderId":  orderId,
	}

	ordersResult, err := api.client.NewUtaBybitServiceWithParams(params).GetOrderHistory(context.Background())
	if err != nil {
		zap.S().Error("Failed to get order history", err)
		return nil, err
	}

	var orderHistory order.OrderHistoryDto
	if err := mapstructure.Decode(ordersResult, &orderHistory); err != nil {
		zap.S().Error("Failed to decode order result", err)
		return nil, err
	}

	return &orderHistory, nil
}

func (api *BybitApi) getSignedApiRequest(uri string, queryParams map[string]interface{}) ([]byte, error) {
	sign := api.getSignature(queryParams)
	url := uri + "?" + util.ConvertMapParamsToString(queryParams) + "&sign=" + sign

	return api.signedApiRequest(http.MethodGet, url, nil)
}

func (api *BybitApi) postSignedApiRequest(uri string, queryParams map[string]interface{}) ([]byte, error) {
	queryParams["sign"] = api.getSignature(queryParams)
	jsonString, _ := json.Marshal(queryParams)

	return api.signedApiRequest(http.MethodPost, uri, bytes.NewBuffer(jsonString))
}

func (api *BybitApi) signedApiRequest(method, uri string, requestBody io.Reader) ([]byte, error) {
	urlRequest := "https://api.bytick.com" + uri
	client := &http.Client{}
	req, err := http.NewRequest(method, urlRequest, requestBody)

	if err != nil {
		zap.S().Errorf("API error: %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

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
	return body, nil
}

func (api *BybitApi) sign(data string) string {
	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(api.secretKey))

	// Write Data to it
	h.Write([]byte(data))

	// Get result and encode as hexadecimal string
	sha := hex.EncodeToString(h.Sum(nil))

	return sha
}

func (api *BybitApi) getSignature(params map[string]interface{}) string {
	h := hmac.New(sha256.New, []byte(api.secretKey))
	io.WriteString(h, util.ConvertMapParamsToString(params))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (api *BybitApi) SetFuturesLeverage(coin *domain.Coin, leverage int) error {
	_, err := api.postSignedApiRequest("/private/linear/position/set-leverage",
		map[string]interface{}{
			"api_key":       api.apiKey,
			"buy_leverage":  strconv.Itoa(leverage),
			"sell_leverage": strconv.Itoa(leverage),
			"symbol":        coin.Symbol,
			"timestamp":     util.MakeTimestamp(),
		},
	)

	return err
}

func (api *BybitApi) SetIsolatedMargin(coin *domain.Coin, leverage int) error {
	_, err := api.postSignedApiRequest("/contract/v3/private/position/switch-isolated",
		map[string]interface{}{
			"api_key":       api.apiKey,
			"tradeMode":     1, //0: cross margin. 1: isolated margin
			"buy_leverage":  strconv.Itoa(leverage),
			"sell_leverage": strconv.Itoa(leverage),
			"symbol":        coin.Symbol,
			"timestamp":     util.MakeTimestamp(),
		},
	)

	return err
}

func (api *BybitApi) OpenFuturesOrder(coin *domain.Coin, amount float64, price float64, futuresType futureType.FuturesType, stopLossPriceInCents float64) (api.OrderResponseDto, error) {
	side := "Buy"
	if futuresType == futureType.SHORT {
		side = "Sell"
	}

	return api.makeFutureOrderByMarket(coin, amount, side)
}

func (api *BybitApi) CloseFuturesOrder(coin *domain.Coin, openedTransaction *domain.Transaction, price float64) (api.OrderResponseDto, error) {
	side := "Sell"
	if openedTransaction.FuturesType == futureType.SHORT {
		side = "Buy"
	}
	return api.makeFutureOrderByMarket(coin, openedTransaction.Amount, side)
}

func (api *BybitApi) makeFutureOrderByMarket(coin *domain.Coin, quantity float64, side string) (api.OrderResponseDto, error) {
	params := map[string]interface{}{
		"category":    "linear",
		"symbol":      coin.Symbol,
		"side":        side,
		"positionIdx": "0",
		"orderType":   "Market",
		"qty":         fmt.Sprintf("%.3f", quantity),
	}
	zap.S().Info("bybit params", params)
	response, err := api.client.NewUtaBybitServiceWithParams(params).PlaceOrder(context.Background())
	if err != nil {
		return nil, err
	}

	dto := order.OrderResponseDto{}
	errDecode := mapstructure.Decode(response, &dto)
	if errDecode != nil {
		return nil, errDecode
	}
	if dto.RetCode != 0 {
		return nil, errors.New(dto.RetMsg)
	}

	for i := 0; i < 60; i++ {
		time.Sleep(time.Second)

		orderHistory, err := api.getOrderById(dto.Result.OrderId)
		if err != nil {
			zap.S().Error("Failed to get order status", err)
			continue
		}

		if orderHistory.Result.List[0].OrderStatus == "Filled" {
			return &orderHistory.Result.List[0], nil
		}
	}

	return nil, errors.New("order not filled after 60 seconds")
}

func (api *BybitApi) futuresOrderByMarket(queryParams map[string]interface{}) (*order.FuturesOrderResponseDto, error) {
	body, err := api.postSignedApiRequest("/private/linear/order/create", queryParams)
	if err != nil {
		return nil, err
	}

	zap.S().Infof("API response: %s", string(body))

	dto := order.FuturesOrderResponseDto{}
	errUnmarshal := json.Unmarshal(body, &dto)
	if errUnmarshal != nil {
		zap.S().Error("Unmarshal error: ", errUnmarshal.Error())
		return nil, errUnmarshal
	}

	if dto.RetCode != 0 {
		return nil, errors.New("create order failed")
	}

	return &dto, nil
}

func (api *BybitApi) IsFuturesPositionOpened(coin *domain.Coin, openedOrder *domain.Transaction) bool {
	positionDto, err := api.GetPosition(coin)
	if err != nil || positionDto.RetCode != 0 {
		zap.S().Error("Error on getting position!")
		return false
	}

	for _, positionDto := range positionDto.Result.List {
		if positionDto.Side == "Buy" && openedOrder.FuturesType == futureType.LONG ||
			positionDto.Side == "Sell" && openedOrder.FuturesType == futureType.SHORT {
			return positionDto.Size != "0"
		}
	}

	return false
}

func (api *BybitApi) GetLastFuturesOrder(coin *domain.Coin, clientOrderId string) (api.OrderResponseDto, error) {
	requestParams := map[string]interface{}{
		"api_key":   api.apiKey,
		"order_id":  clientOrderId,
		"timestamp": util.MakeTimestamp(),
		"symbol":    coin.Symbol,
	}

	body, err := api.getSignedApiRequest("/private/linear/order/list", requestParams)
	if err != nil {
		return nil, err
	}

	dto := order.ActiveOrdersResponseDto{}
	errUnmarshal := json.Unmarshal(body, &dto)
	if errUnmarshal != nil {
		zap.S().Error("Unmarshal error", errUnmarshal.Error())
		return nil, errUnmarshal
	}

	if len(dto.Result.Data) > 0 {
		return &dto.Result.Data[0], nil
	}

	return nil, nil
}

func (api *BybitApi) GetFuturesActiveOrdersByCoin(coin *domain.Coin) (*order.ActiveOrdersResponseDto, error) {
	requestParams := map[string]interface{}{
		"api_key":   api.apiKey,
		"timestamp": util.MakeTimestamp(),
		"symbol":    coin.Symbol,
	}

	body, err := api.getSignedApiRequest("/private/linear/order/list", requestParams)
	if err != nil {
		return nil, err
	}

	dto := order.ActiveOrdersResponseDto{}
	errUnmarshal := json.Unmarshal(body, &dto)
	if errUnmarshal != nil {
		zap.S().Error("Unmarshal error", errUnmarshal.Error())
		return nil, errUnmarshal
	}

	return &dto, nil
}

func (api *BybitApi) GetActiveOrder(orderDto *order.FuturesOrderResponseDto) (api.OrderResponseDto, error) {
	requestParams := map[string]interface{}{
		"api_key":   api.apiKey,
		"order_id":  orderDto.Result.OrderId,
		"timestamp": util.MakeTimestamp(),
		"symbol":    orderDto.Result.Symbol,
	}

	body, err := api.getSignedApiRequest("/private/linear/order/list", requestParams)
	if err != nil {
		return nil, err
	}

	dto := order.ActiveOrdersResponseDto{}
	errUnmarshal := json.Unmarshal(body, &dto)
	if errUnmarshal != nil {
		zap.S().Error("Unmarshal error", errUnmarshal.Error())
		return nil, errUnmarshal
	}

	if len(dto.Result.Data) == 0 {
		return nil, errors.New("empty response")
	}

	return &dto.Result.Data[0], nil
}

func (api *BybitApi) GetWalletBalance() (api.WalletBalanceDto, error) {
	params := map[string]interface{}{
		"accountType": "UNIFIED",
		"coin":        "USDT",
	}

	result, err := api.client.NewUtaBybitServiceWithParams(params).GetAccountWallet(context.Background())
	if err != nil {
		return nil, err
	}
	zap.S().Debug("GetWalletBalance", result)

	dto := wallet.GetWalletBalanceDto{}
	if err := mapstructure.Decode(result, &dto); err != nil {
		zap.S().Error("Failed to decode order result", err)
		return nil, err
	}

	return &dto, nil
}

func (api *BybitApi) GetConditionalOrder(coin *domain.Coin) (*order.GetConditionalOrderDto, error) {
	requestParams := map[string]interface{}{
		"api_key":   api.apiKey,
		"timestamp": util.MakeTimestamp(),
		"symbol":    coin.Symbol,
	}

	body, err := api.getSignedApiRequest("/private/linear/stop-order/list", requestParams)
	if err != nil {
		return nil, err
	}

	dto := order.GetConditionalOrderDto{}
	errUnmarshal := json.Unmarshal(body, &dto)
	if errUnmarshal != nil {
		zap.S().Error("Unmarshal error", errUnmarshal.Error())
		return nil, errUnmarshal
	}

	return &dto, nil
}

func (api *BybitApi) GetPosition(coin *domain.Coin) (*position.GetPositionDto, error) {
	params := map[string]interface{}{"category": "linear", "symbol": coin.Symbol, "limit": 1}
	response, err := api.client.NewUtaBybitServiceWithParams(params).GetPositionList(context.Background())
	if err != nil {
		return nil, err
	}

	dto := position.GetPositionDto{}
	errDecode := mapstructure.Decode(response, &dto)
	if errDecode != nil {
		return nil, errDecode
	}

	return &dto, nil
}

func (api *BybitApi) GetTradeRecords(coin *domain.Coin, openTransaction *domain.Transaction) (*position.GetTradeRecordsDto, error) {
	requestParams := map[string]interface{}{
		"api_key":    api.apiKey,
		"symbol":     coin.Symbol,
		"exec_type":  "Trade",
		"start_time": util.GetMillisByTime(openTransaction.CreatedAt),
		"timestamp":  util.MakeTimestamp(),
	}

	body, err := api.getSignedApiRequest("/private/linear/trade/execution/list", requestParams)
	if err != nil {
		return nil, err
	}

	dto := position.GetTradeRecordsDto{}
	errUnmarshal := json.Unmarshal(body, &dto)
	if errUnmarshal != nil {
		zap.S().Error("Unmarshal error", errUnmarshal.Error())
		return nil, errUnmarshal
	}

	return &dto, nil
}

func (api *BybitApi) GetCloseTradeRecord(coin *domain.Coin, openTransaction *domain.Transaction) (api.OrderResponseDto, error) {
	tradeRecordsDto, err := api.GetTradeRecords(coin, openTransaction)
	if err != nil {
		return nil, err
	}

	var trades []position.TradeRecordDto

	for _, tradeRecordDto := range tradeRecordsDto.Result.Data {
		if tradeRecordDto.Side == "Sell" && openTransaction.FuturesType == futureType.LONG ||
			tradeRecordDto.Side == "Buy" && openTransaction.FuturesType == futureType.SHORT {
			trades = append(trades, tradeRecordDto)
		}
	}

	tradesSummaryDto := position.TradesSummaryDto{Trades: trades}

	if tradesSummaryDto.GetAmount() != openTransaction.Amount {
		error := fmt.Sprintf("Unexpected amount in trade records. Expected: %v; actual: %v", openTransaction.Amount, tradesSummaryDto.GetAmount())
		zap.S().Error(error)
		return nil, errors.New(error)
	}

	return &tradesSummaryDto, nil
}

func (api *BybitApi) ReplaceFuturesActiveOrder(coin *domain.Coin, transaction *domain.Transaction, stopLossPriceInCents int64) (*order.ReplaceFuturesActiveOrder, error) {
	queryParams := map[string]interface{}{
		"api_key":   api.apiKey,
		"order_id":  transaction.ClientOrderId.String,
		"symbol":    coin.Symbol,
		"stop_loss": util.GetDollarsByCents(stopLossPriceInCents),
		"timestamp": util.MakeTimestamp(),
	}

	body, err := api.postSignedApiRequest("/private/linear/order/replace", queryParams)
	if err != nil {
		return nil, err
	}

	dto := order.ReplaceFuturesActiveOrder{}
	errUnmarshal := json.Unmarshal(body, &dto)
	if errUnmarshal != nil {
		zap.S().Error("Unmarshal error: ", errUnmarshal.Error())
		return nil, errUnmarshal
	}

	if dto.RetCode != 0 {
		return nil, errors.New(dto.RetMsg)
	}

	return &dto, nil
}

func (api *BybitApi) SetApiKey(apiKey string) {
	api.apiKey = apiKey
}

func (api *BybitApi) SetSecretKey(secretKey string) {
	api.secretKey = secretKey
}
