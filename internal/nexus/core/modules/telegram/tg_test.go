package telegram

import (
	"DaruBot/internal/config"
	"DaruBot/internal/nexus/core"
	"DaruBot/pkg/logger"
	"DaruBot/pkg/nexus"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

func newTgWh() (*TgBot, error) {
	cfg := config.GetDefaultConfig()
	cfg.Nexus.TLS = config.TLSCert{
		Url:      "",
		KeyFile:  "../../../../assets/certs/key.pem",
		CertFile: "../../../../assets/certs/cert.pem",
	}
	//config.Config.Nexus.Proxy.Addr = "94.130.73.18:1145"
	cfg.Nexus.Modules.Telegram.WebhookMode = true
	cfg.Nexus.Modules.Telegram.CustomCert = true

	lg := logger.New(os.Stdout, logger.TraceLevel)

	return NewTelegram(cfg, lg)
}

func newTg() (*TgBot, error) {
	lg := logger.New(os.Stdout, logger.TraceLevel)

	return NewTelegram(config.GetDefaultConfig(), lg)
}

func dumbHandler(ctx context.Context, cmd nexus.Command) (nexus.Response, error) {
	return &core.Response{
		Rsp: nil,
	}, nil
}

func Test_TLS(t *testing.T) {
	// openssl req -newkey rsa:2048 -sha256 -nodes -keyout key.pem -x509 -days 365 -out cert.pem -addext 'subjectAltName = IP:127.0.0.1' -subj '/C=US/ST=CA/L=SanFrancisco/O=MyCompany/OU=RND/CN=127.0.0.1/'

	helloServer := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, err := w.Write([]byte("This is an example server.\n"))
		if err != nil {
			fmt.Printf("Write: %v\n", err)
		}
	}

	http.HandleFunc("/hello", helloServer)

	go func() {
		err := http.ListenAndServeTLS(":8443", "../../../../assets/certs/cert.pem", "../../../../assets/certs/key.pem", nil)
		if err != nil {
			t.Fatal("ListenAndServe: ", err)
		}
	}()

	time.Sleep(1 * time.Second)

	cert, err := ioutil.ReadFile("../../../../assets/certs/cert.pem")
	if err != nil {
		t.Fatalf("Couldn't load file %v", err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(cert)

	tlsConf := &tls.Config{RootCAs: certPool, InsecureSkipVerify: true}
	tr := &http.Transport{TLSClientConfig: tlsConf} /*DialContext: (&tls.Dialer{
		NetDialer: &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
		Config: tlsConf,
	}).DialContext*/
	client := &http.Client{Transport: tr}

	resp, err := client.Get("https://127.0.0.1:8443/hello")
	if err != nil {
		t.Fatalf("Couldn't get %v", err)
	}
	defer resp.Body.Close()

	t.Log(resp.Status)

}

func Test_TelegramWebHook(t *testing.T) {
	// openssl req -newkey rsa:2048 -sha256 -nodes -keyout key.pem -x509 -days 365 -out cert.pem -subj '/C=US/ST=CA/L=SanFrancisco/O=MyCompany/OU=RND/CN={IP or HOST}'
	// curl -F "url=https://{IP or HOST}:8443/" -F "certificate=@cert.pem" "https://api.telegram.org/bot{TOKEN}/setwebhook"

	tg, err := newTgWh()
	if err != nil {
		t.Fatal(err)
	}

	err = tg.Init(dumbHandler)
	if err != nil {
		t.Fatal(err)
	}
	defer tg.Stop()

	stop := false

	tg.bot.Handle(tb.OnText, func(m *tb.Message) {
		t.Logf("[%s] recive text: %#v", time.Now().Format("15:04:05"), m.Text)
		if m.Text == "/exit" {
			stop = true
		}
		_, err = tg.bot.Send(m.Sender, m.Text)
		if err != nil {
			t.Fatal(err)
		}
	})

	for !stop {
	}
}

func Test_Telegram(t *testing.T) {
	tg, err := newTg()
	if err != nil {
		t.Fatal(err)
	}

	err = tg.Init(dumbHandler)
	if err != nil {
		t.Fatal(err)
	}
	defer tg.Stop()

	stop := false

	tg.bot.Handle(tb.OnText, func(m *tb.Message) {
		t.Logf("[%s] recive text: %#v", time.Now().Format("15:04:05"), m.Text)
		if m.Text == "/exit" {
			stop = true
		}
		_, err = tg.bot.Send(m.Sender, m.Text)
		if err != nil {
			t.Fatal(err)
		}
	})

	for !stop {
	}
}

func Test_TelegramSendToGroup(t *testing.T) {
	tg, err := newTg()
	if err != nil {
		t.Fatal(err)
	}

	err = tg.Init(dumbHandler)
	if err != nil {
		t.Fatal(err)
	}
	defer tg.Stop()

	_, err = tg.bot.Send(tb.ChatID(tg.cfg.Nexus.Modules.Telegram.GroupID), fmt.Sprintf("@%s 123", tg.userName))
	if err != nil {
		t.Fatal(err)
	}
	_, err = tg.bot.Send(tb.ChatID(tg.cfg.Nexus.Modules.Telegram.GroupID), "123", tb.Silent, tb.NoPreview)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tg.bot.Send(tb.ChatID(tg.cfg.Nexus.Modules.Telegram.GroupID), "`123`", tb.ModeMarkdownV2)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_TelegramSendMenu(t *testing.T) {
	tg, err := newTg()
	if err != nil {
		t.Fatal(err)
	}

	err = tg.Init(dumbHandler)
	if err != nil {
		t.Fatal(err)
	}
	defer tg.Stop()

	menu := &tb.ReplyMarkup{ResizeReplyKeyboard: true}

	btnHelp := menu.Text("1")
	btnSettings := menu.Text("2")

	//menu.Text("Hello!")
	//menu.Contact("Send phone number")
	//menu.Location("Send location")
	//menu.Poll(tb.PollQuiz)

	menu.Reply(
		menu.Row(btnHelp),
		menu.Row(btnSettings),
	)

	// b.Handle(&btnHelp, func(m *tb.Message) {...})

	_, err = tg.bot.Send(tb.ChatID(tg.cfg.Nexus.Modules.Telegram.UserID), "Hello!", menu)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_TelegramSendSelector(t *testing.T) {
	tg, err := newTg()
	if err != nil {
		t.Fatal(err)
	}

	err = tg.Init(dumbHandler)
	if err != nil {
		t.Fatal(err)
	}
	defer tg.Stop()

	selector := &tb.ReplyMarkup{}

	btnPrev := selector.Data("⬅", "prev")
	btnNext := selector.Data("➡", "next")

	// Inline buttons:
	//selector.Data("Show help", "help") // data is optional
	//selector.Data("Delete item", "delete", item.ID)
	//selector.URL("Visit", "https://google.com")
	//selector.Query("Search", query)
	//selector.QueryChat("Share", query)
	//selector.Login("Login", &tb.Login{...})

	selector.Inline(
		selector.Row(btnPrev, btnNext),
	)

	//b.Handle(&btnPrev, func(c *tb.Callback) {
	//	// ...
	//	// Always respond!
	//	b.Respond(c, &tb.CallbackResponse{...})
	//})

	_, err = tg.bot.Send(tb.ChatID(tg.cfg.Nexus.Modules.Telegram.UserID), "Hello!", selector)
	if err != nil {
		t.Fatal(err)
	}
}
