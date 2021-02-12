#!/bin/sh

resolution="d"

# https://github.com/markcheno/go-quote
./quote -start=2021-02-08 -period=$resolution -delay=300 -outfile=BTC_USD_$resolution.csv -source=coinbase BTC-USD