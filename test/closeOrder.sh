#!/bin/bash

# Check if all required arguments are provided
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <symbol> <strategyTag>"
    echo "Example: $0 BTCUSDT ChatGpt"
    exit 1
fi

SYMBOL=$1
STRATEGY_TAG=$2

# Make the request to the debug endpoint
curl -X GET "http://localhost:8080/debug/close/${SYMBOL}/${STRATEGY_TAG}"
echo ""