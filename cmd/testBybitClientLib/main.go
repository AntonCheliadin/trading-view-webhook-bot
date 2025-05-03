package main

import (
	"context"
	"fmt"
	bybit "github.com/bybit-exchange/bybit.go.api"
	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"os"
	"tradingViewWebhookBot/internal/dto/bybit/order"
	"tradingViewWebhookBot/internal/dto/bybit/wallet"
	"tradingViewWebhookBot/internal/logger"
)

func main() {
	// Initialize logger
	logger := logger.InitLogger()
	defer logger.Sync()

	logger.Info("testBybitClientLib starting...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Fatal("Error loading .env file", zap.Error(err))
	}

	//getPrice()
	//buyOrder()
	//getPosition()
	//sellOrder()
	//getOrders()
	//getOrderById()
	//getWalletBalance()
}

func getPrice() {
	params := map[string]interface{}{
		"category": "linear", // Important: "linear" = USDT perpetual
		"symbol":   "XRPUSDT",
		"interval": "1",
		"limit":    "1",
	}

	client := bybit.NewBybitHttpClient(os.Getenv("BYBIT_API_KEY"), os.Getenv("BYBIT_API_SECRET"), bybit.WithBaseURL(bybit.TESTNET))

	priceResult, err := client.NewUtaBybitServiceWithParams(params).GetMarkPriceKline(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(bybit.PrettyPrint(priceResult.Result))
}

func buyOrder() {
	client := bybit.NewBybitHttpClient(os.Getenv("BYBIT_API_KEY"), os.Getenv("BYBIT_API_SECRET"), bybit.WithBaseURL(bybit.MAINNET))

	params := map[string]interface{}{
		"category":    "linear",
		"symbol":      "BTCUSDT",
		"side":        "Buy",
		"positionIdx": 0,
		"orderType":   "Market",
		"qty":         "0.001",
	}
	response, err := client.NewUtaBybitServiceWithParams(params).PlaceOrder(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("-----------" + "PlaceOrder BUY" + "-----------")
	fmt.Println(bybit.PrettyPrint(response))
}

func sellOrder() {
	client := bybit.NewBybitHttpClient(os.Getenv("BYBIT_API_KEY"), os.Getenv("BYBIT_API_SECRET"), bybit.WithBaseURL(bybit.MAINNET))

	params := map[string]interface{}{
		"category":    "linear",
		"symbol":      "XRPUSDT",
		"side":        "Sell",
		"positionIdx": 0,
		"orderType":   "Market",
		"qty":         "8",
	}
	response, err := client.NewUtaBybitServiceWithParams(params).PlaceOrder(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("-----------" + "PlaceOrder SELL" + "-----------")
	fmt.Println(bybit.PrettyPrint(response))
}

func getPosition() {
	client := bybit.NewBybitHttpClient(os.Getenv("BYBIT_API_KEY"), os.Getenv("BYBIT_API_SECRET"), bybit.WithBaseURL(bybit.MAINNET))

	params := map[string]interface{}{"category": "linear", "settleCoin": "USDT", "limit": 100}
	orderResult, err := client.NewUtaBybitServiceWithParams(params).GetPositionList(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("-----------" + "getPosition" + "-----------")
	fmt.Println(bybit.PrettyPrint(orderResult))
}

func getOrders() {
	client := bybit.NewBybitHttpClient(os.Getenv("BYBIT_API_KEY"), os.Getenv("BYBIT_API_SECRET"), bybit.WithBaseURL(bybit.MAINNET))

	params := map[string]interface{}{"category": "linear", "symbol": "XRPUSDT", "limit": 10}
	ordersResult, err := client.NewUtaBybitServiceWithParams(params).GetOrderHistory(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("-----------" + "getOrders" + "-----------")
	fmt.Println(bybit.PrettyPrint(ordersResult))
}

func getOrderById() {
	client := bybit.NewBybitHttpClient(os.Getenv("BYBIT_API_KEY"), os.Getenv("BYBIT_API_SECRET"), bybit.WithBaseURL(bybit.MAINNET))

	params := map[string]interface{}{
		"category": "linear",
		"orderId":  "cf8846d6-7492-4592-a1c0-68d5af5a51e2",
	}

	ordersResult, err := client.NewUtaBybitServiceWithParams(params).GetOrderHistory(context.Background())
	if err != nil {
		zap.S().Error("Failed to get order history", err)
		return
	}

	var orderHistory order.OrderHistoryDto
	if err := mapstructure.Decode(ordersResult, &orderHistory); err != nil {
		zap.S().Error("Failed to decode order result", err)
		return
	}

	fmt.Println("-----------" + "getOrderById" + "-----------")
	fmt.Println(bybit.PrettyPrint(orderHistory))

	return

}

func getWalletBalance() {
	client := bybit.NewBybitHttpClient(os.Getenv("BYBIT_API_KEY"), os.Getenv("BYBIT_API_SECRET"), bybit.WithBaseURL(bybit.MAINNET))

	params := map[string]interface{}{
		"accountType": "UNIFIED",
		"coin":        "USDT",
	}

	result, err := client.NewUtaBybitServiceWithParams(params).GetAccountWallet(context.Background())
	if err != nil {
		zap.S().Error("Failed to get order history", err)
		return
	}

	dto := wallet.GetWalletBalanceDto{}
	if err := mapstructure.Decode(result, &dto); err != nil {
		zap.S().Error("Failed to decode order result", err)
		return
	}

	fmt.Println("-----------" + "getWalletBalance" + "-----------")
	fmt.Println(bybit.PrettyPrint(dto))

	fmt.Println("-----------" + "getWalletBalance GetAvailableBalance" + "-----------")
	fmt.Println(dto.GetAvailableBalance())

}
