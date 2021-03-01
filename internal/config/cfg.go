package config

import (
	"DaruBot/pkg/tools"
	"os"
)

type Configurations struct {
	debugMode bool
	Storage   Storage

	Logger     Logger
	Exchanges  Exchanges
	Strategies map[string]interface{}
	Nexus      Nexus
}

type Logger struct {
	FileOutput bool
}

type Exchanges struct {
	Bitfinex Bitfinex
}

type Bitfinex struct {
	ApiKey    string `mapstructure:",omitempty" yaml:",omitempty"`
	ApiSec    string `mapstructure:",omitempty" yaml:",omitempty"`
	Strategy  string
	affiliate string
}

func (b *Bitfinex) Affiliate() string {
	return b.affiliate
}

type DaruStonks struct {
	Pair   string
	Margin bool
}

type Nexus struct {
	Modules Modules
	TLS     TLSCert
	Proxy   Proxy
}

type Modules struct {
	Telegram Telegram
}

type Telegram struct {
	Enabled     bool
	WebhookMode bool
	CustomCert  bool
	APIKey      string `mapstructure:",omitempty" yaml:",omitempty"`
	GroupID     int    `mapstructure:",omitempty" yaml:",omitempty"`
	UserID      int    `mapstructure:",omitempty" yaml:",omitempty"`
}

type TLSCert struct {
	Url      string `mapstructure:",omitempty" yaml:",omitempty"`
	KeyFile  string
	CertFile string
}

type Proxy struct {
	Addr string `mapstructure:",omitempty" yaml:",omitempty"`
}

type Storage struct {
	Local StorageLocal
}

type StorageLocal struct {
	Path string
}

var (
	defaultConfig = Configurations{
		debugMode: true,
		Logger: Logger{
			FileOutput: false,
		},
		Exchanges: Exchanges{
			Bitfinex: Bitfinex{
				ApiKey:    "",
				ApiSec:    "",
				Strategy:  "",
				affiliate: "jXAX6tEPA",
			},
		},
		Strategies: make(map[string]interface{}),
		Nexus: Nexus{
			Modules: Modules{
				Telegram{
					Enabled:     false,
					WebhookMode: false,
					CustomCert:  false,
					APIKey:      "",
					GroupID:     0,
					UserID:      0,
				},
			},
			TLS: TLSCert{
				Url:      "",
				KeyFile:  "",
				CertFile: "",
			},
			Proxy: Proxy{
				Addr: "",
			},
		},
		Storage: Storage{
			Local: StorageLocal{
				Path: "./storage.db",
			},
		},
	}
)

func (c *Configurations) IsDebug() bool {
	return c.debugMode
}

func (c *Configurations) SetDebug(enable bool) {
	c.debugMode = enable
}

func GetDefaultConfig() Configurations {
	cfg := defaultConfig

	cfg.Exchanges.Bitfinex.ApiKey = os.Getenv("BITFINEX_API_KEY")
	cfg.Exchanges.Bitfinex.ApiSec = os.Getenv("BITFINEX_API_SEC")

	cfg.Nexus.Modules.Telegram.APIKey = os.Getenv("TG_API_KEY")
	cfg.Nexus.Modules.Telegram.GroupID = tools.StrToIntMust(os.Getenv("TG_GROUP_ID"))
	cfg.Nexus.Modules.Telegram.UserID = tools.StrToIntMust(os.Getenv("TG_USER_ID"))

	cfg.Nexus.TLS.Url = os.Getenv("TLS_HOST")

	cfg.Nexus.Proxy.Addr = os.Getenv("PROXY_ADDR")

	return cfg
}
