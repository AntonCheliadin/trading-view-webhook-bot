#!/bin/bash

# Check if all required arguments are provided
if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <symbol> <strategyTag> <futureType>"
    echo "Example: $0 BTCUSDT ChatGpt 0"
    echo "futureType: 0 for LONG, 1 for SHORT"
    exit 1
fi

SYMBOL=$1
STRATEGY_TAG=$2
FUTURE_TYPE=$3

# Make the request to the debug endpoint
curl -X GET "http://localhost:8080/debug/open/${SYMBOL}/${STRATEGY_TAG}/${FUTURE_TYPE}"
echo ""