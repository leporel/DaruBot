package telegram

import (
	"DaruBot/internal/config"
	"DaruBot/pkg/logger"
	"DaruBot/pkg/nexus"
	"DaruBot/pkg/proxy"
	"DaruBot/pkg/tools"
	"encoding/json"
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"strconv"
	"time"
)

const (
	NexusModuleTelegram nexus.ModuleName = "Telegram"
)

type TgBot struct {
	bot *tb.Bot
	log logger.Logger
	cfg config.Configurations
}

func NewTelegram(cfg config.Configurations, lg logger.Logger) (*TgBot, error) {
	tlog := lg.WithPrefix("nexus", NexusModuleTelegram)

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

		poller = &tb.Webhook{
			Listen:   ":8443",
			Endpoint: endp,
			TLS:      whTLS,
			// Not necessary, used to log after
			HasCustomCert: cfg.Nexus.Modules.Telegram.CustomCert,
		}
	}

	httpCli, err := proxy.NewProxyClient(cfg.Nexus.Proxy.Addr)
	if err != nil {
		return nil, err
	}

	rp := func(err error) {
		tlog.Error(err)
	}

	b, err := tb.NewBot(tb.Settings{
		Client:      httpCli,
		Token:       cfg.Nexus.Modules.Telegram.APIKey,
		Poller:      poller,
		Reporter:    rp,
		Verbose:     true,
		Synchronous: true,
	})
	if err != nil {
		return nil, err
	}

	return &TgBot{
		bot: b,
		log: tlog,
		cfg: cfg,
	}, nil
}

func (t *TgBot) Init() error {
	if !t.cfg.Nexus.Modules.Telegram.WebhookMode {
		if err := t.startLongPolling(); err != nil {
			return err
		}
	} else {
		if err := t.startWebhook(); err != nil {
			return err
		}

	}

	// TODO
	//
	// check can_join_groups

	time.Sleep(1 * time.Second)
	t.log.Info("Bot now listening commands")

	return nil
}

func (t *TgBot) startLongPolling() error {
	t.log.Debug("web hook deleted")
	if err := t.bot.RemoveWebhook(); err != nil {
		return err
	}

	skipped := 0

	for {
		offset := t.bot.Poller.(*tb.LongPoller).LastUpdateID + 1

		params := map[string]string{
			"offset":  strconv.Itoa(offset),
			"timeout": strconv.Itoa(5),
		}
		data, err := t.bot.Raw("getUpdates", params)
		if err != nil {
			return err
		}
		var resp struct {
			Result []tb.Update
		}
		if err = json.Unmarshal(data, &resp); err != nil {
			return err
		}
		if len(resp.Result) == 0 {
			break
		}

		for _, update := range resp.Result {
			t.bot.Poller.(*tb.LongPoller).LastUpdateID = update.ID
			skipped++
		}

		time.Sleep(2 * time.Second)
	}

	t.log.Tracef("last upd ID: %v", t.bot.Poller.(*tb.LongPoller).LastUpdateID)
	t.log.Debugf("skipped %v updates", skipped)

	go t.bot.Start()

	return nil
}

func (t *TgBot) startWebhook() error {
	wh := t.bot.Poller.(*tb.Webhook)

	t.log.Infof("Telegram webhook listening address %s. Used custom cert: %v\n",
		wh.Endpoint.PublicURL,
		wh.HasCustomCert)

	wh, err := t.bot.GetWebhook()
	if err != nil {
		return err
	}
	pending := wh.PendingUpdates

	skipped := 0

	if pending > 0 {
		skipDone := make(chan struct{}, 1)
		updates := make(chan tb.Update, 10)
		go t.bot.Poller.Poll(t.bot, updates, skipDone)

		for pending > skipped {
			select {
			case _ = <-updates:
				skipped++
			}
		}
		skipDone <- struct{}{}
		defer close(updates)

		time.Sleep(2 * time.Second)
	}

	t.log.Debugf("skipped %v updates", skipped)

	go t.bot.Start()

	return nil
}

func (t *TgBot) Stop() error {
	t.log.Info("Stopping bot")

	t.bot.Stop()

	return nil
}

/*func (n *nexus) Fire(hd *logger.HookData) error {

	msg := &Notification{
		Msg: fmt.Sprintf("[%s] [%s]\n%s",
			hd.Level, hd.Time.Format("01.02 15:04:05"), hd.Message),
		Raw: hd,
	}

	switch {
	case hd.Level > logger.WarnLevel:
		msg.Kind = NotifyKindLog
	case hd.Level == logger.WarnLevel:
		msg.Kind = NotifyKindWarning
	default:
		msg.Kind = NotifyKindError
	}

	return n.Send(msg)
}*/
