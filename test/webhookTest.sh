curl -X POST \
-H 'Content-Type: application/json; charset=utf-8' \
-d '{
    "text": "ChatGpt Trading strategy: order {{strategy.order.action}} @ {{strategy.order.contracts}} filled on {{ticker}}.",
    "side": "buy",
    "ticker": "BTCUSDT",
    "tag": "ChatGpt",
    "interval": "1D",
    "price": "0.1809",
    "positionSize": ""
}' \
http://localhost:8081/webhook/alert

