package config

import (
	"DaruBot/cmd"
	"DaruBot/pkg/tools"
	"os"
)

type Configurations struct {
	debugMode  bool
	Logger     Logger
	Exchanges  Exchanges
	Strategies Strategies
	Nexus      Nexus
}

var (
	Config = Configurations{
		debugMode: cmd.DebugMode,
		Logger: Logger{
			FileOutput: false,
		},
		Exchanges: Exchanges{
			Bitfinex: Bitfinex{
				ApiKey:    os.Getenv("BITFINEX_API_KEY"),
				ApiSec:    os.Getenv("BITFINEX_API_SEC"),
				Strategy:  "DaruStonks",
				Affiliate: "jXAX6tEPA",
			},
		},
		Strategies: Strategies{
			DaruStonks: DaruStonks{
				Pair:   "tTESTBTC:TESTUSD",
				Margin: false,
			}},
		Nexus: Nexus{
			Modules: Modules{
				Telegram{
					Enabled:     false,
					WebhookMode: false,
					CustomCert:  false,
					APIKey:      os.Getenv("TG_API_KEY"),
					GroupID:     tools.StrToIntMust(os.Getenv("TG_GROUP_ID")),
					UserID:      tools.StrToIntMust(os.Getenv("TG_USER_ID")),
				},
			},
			TLSCert: TLSCert{
				Url:      os.Getenv("TLS_HOST"),
				KeyFile:  "",
				CertFile: "",
			},
			Proxy: Proxy{
				Addr: os.Getenv("PROXY_ADDR"),
			},
		},
	}
)

func (c Configurations) IsDebug() bool {
	return c.debugMode
}

type Logger struct {
	FileOutput bool
}

type Exchanges struct {
	Bitfinex Bitfinex
}

type Bitfinex struct {
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

type Nexus struct {
	Modules Modules
	TLSCert TLSCert
	Proxy   Proxy
}

type Modules struct {
	Telegram Telegram
}

type Telegram struct {
	Enabled     bool
	WebhookMode bool
	CustomCert  bool
	APIKey      string
	GroupID     int
	UserID      int
}

type TLSCert struct {
	Url      string
	KeyFile  string
	CertFile string
}

type Proxy struct {
	Addr string
}
