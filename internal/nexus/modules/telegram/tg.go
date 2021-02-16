package telegram

import (
	"DaruBot/internal/config"
	"DaruBot/internal/models"
	"DaruBot/internal/nexus"
	"DaruBot/pkg/tools"
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"time"
)

const (
	NexusModuleTelegram models.NexusModuleName = "Telegram"
)

type TgBot struct {
	bot *tb.Bot
}

func NewTelegram(cfg config.Configurations) (*TgBot, error) {

	var poller tb.Poller = &tb.LongPoller{
		Timeout: 15 * time.Second,
	}

	if cfg.Nexus.Modules.Telegram.WebhookMode {
		host := cfg.Nexus.TLSCert.Url

		if host == "" {
			myIP, err := tools.GetIP(nil)
			if err != nil {
				return nil, err
			}
			host = fmt.Sprintf("https://%s:8443/", myIP)
		}

		var whTLS *tb.WebhookTLS
		endp := &tb.WebhookEndpoint{
			PublicURL: host,
		}

		// If self signed cert provided
		if cfg.Nexus.Modules.Telegram.CustomCert {
			whTLS = &tb.WebhookTLS{
				Key:  cfg.Nexus.TLSCert.KeyFile,
				Cert: cfg.Nexus.TLSCert.CertFile,
			}
			endp.Cert = cfg.Nexus.TLSCert.CertFile
		}

		fmt.Printf("Telegram webhook listening IP: %s. Used custom cert: %v\n",
			host,
			cfg.Nexus.Modules.Telegram.CustomCert)

		poller = &tb.Webhook{
			Listen:   ":8443",
			Endpoint: endp,
			TLS:      whTLS,
		}
	}

	httpCli, err := nexus.NewProxyClient(cfg.Nexus.Proxy.Addr)
	if err != nil {
		return nil, err
	}

	rp := func(err error) {
		// TODO change to log
		fmt.Println(err)
	}

	b, err := tb.NewBot(tb.Settings{
		Client:   httpCli,
		Token:    cfg.Nexus.Modules.Telegram.APIKey,
		Poller:   poller,
		Reporter: rp,
		Verbose:  true,
	})
	if err != nil {
		return nil, err
	}

	if _, ok := poller.(*tb.Webhook); !ok {
		fmt.Println("web hook deleted")
		err = b.RemoveWebhook()
		if err != nil {
			return nil, err
		}
	}

	fmt.Println("Starting bot")

	// TODO
	// skip pending_update_count (getWebhookInfo)
	// check can_join_groups

	go func() {
		b.Start()
	}()

	return &TgBot{
		bot: b,
	}, nil
}
