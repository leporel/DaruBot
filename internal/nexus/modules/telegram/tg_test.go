package telegram

import (
	"DaruBot/internal/config"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func newTgWh() (*TgBot, error) {
	config.Config.Nexus.TLSCert = config.TLSCert{
		Url:      "",
		KeyFile:  "../../../../assets/certs/key.pem",
		CertFile: "../../../../assets/certs/cert.pem",
	}
	//config.Config.Nexus.Proxy.Addr = "94.130.73.18:1145"
	config.Config.Nexus.Modules.Telegram.WebhookMode = true
	config.Config.Nexus.Modules.Telegram.CustomCert = true
	return NewTelegram(config.Config)
}

func newTg() (*TgBot, error) {
	return NewTelegram(config.Config)
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, err := w.Write([]byte("This is an example server.\n"))
	if err != nil {
		fmt.Printf("Write: %v\n", err)
	}
}

func Test_TLS(t *testing.T) {
	// openssl req -newkey rsa:2048 -sha256 -nodes -keyout key.pem -x509 -days 365 -out cert.pem -addext 'subjectAltName = IP:127.0.0.1' -subj '/C=US/ST=CA/L=SanFrancisco/O=MyCompany/OU=RND/CN=127.0.0.1/'

	http.HandleFunc("/hello", HelloServer)

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

	stop := false

	ws, err := tg.bot.GetWebhook()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("\nwsLastTime: %v\nwsMsg:%v\nCustomCert:%v\nPending:%v\n",
		ws.ErrorUnixtime,
		ws.ErrorMessage,
		ws.HasCustomCert,
		ws.PendingUpdates)

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
