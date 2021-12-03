#!/bin/sh

resolution="5m"

# https://github.com/markcheno/go-quote
./quote -start=2021-02-20 -period=$resolution -delay=300 -outfile=../test_data/BTC_USD_$resolution.csv -source=coinbase BTC-USD