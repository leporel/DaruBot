package proxy

import (
	"DaruBot/internal/config"
	"DaruBot/pkg/tools/network"
	"testing"
)

func Test_ProxyEmpty(t *testing.T) {
	cli, err := NewProxyClient("")
	if err != nil {
		t.Fatal(err)
	}

	ip, err := network.GetIP(cli)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ip)
}

func Test_ProxyAddr(t *testing.T) {
	_, err := NewProxyClient("user:pass@127.0.0.1:777")
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ProxyError(t *testing.T) {
	_, err := NewProxyClient(":pass@127.0.0.1:777")
	if err == nil {
		t.Fatal(err)
	}
}

func Test_Proxy(t *testing.T) {
	cfg := config.GetDefaultConfig()
	if cfg.Nexus.Proxy.Addr == "" {
		t.SkipNow()
	}

	cli, err := NewProxyClient(cfg.Nexus.Proxy.Addr)
	if err != nil {
		t.Fatal(err)
	}

	ip, err := network.GetIP(cli)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ip)
}
