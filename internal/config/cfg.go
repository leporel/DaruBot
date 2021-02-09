package config

import (
	"DaruBot/cmd"
	"os"
)

type Configurations struct {
	debugMode  bool
	Logger     Logger
	Exchanges  Exchanges
	Strategies Strategies
}

var (
	Config = Configurations{
		debugMode: cmd.DebugMode,
		Logger: Logger{
			FileOutput: false,
		},
		Exchanges: Exchanges{
			BitFinex: BitFinex{
				ApiKey:    os.Getenv("API_KEY"),
				ApiSec:    os.Getenv("API_SEC"),
				Strategy:  "DaruStonks",
				Affiliate: "jXAX6tEPA",
			},
		},
		Strategies: Strategies{
			DaruStonks: DaruStonks{
				Pair:   "tTESTBTC:TESTUSD",
				Margin: false,
			}},
	}
)

func (c Configurations) IsDebug() bool {
	return c.debugMode
}

type Logger struct {
	FileOutput bool
}

type Exchanges struct {
	BitFinex BitFinex
}

type BitFinex struct {
	ApiKey    string
	ApiSec    string
	Strategy  string
	Affiliate string
}

type DaruStonks struct {
	Pair   string
	Margin bool
}

type Strategies struct {
	DaruStonks DaruStonks
}
