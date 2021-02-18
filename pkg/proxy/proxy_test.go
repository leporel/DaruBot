package proxy

import (
	"DaruBot/pkg/tools"
	"testing"
)

func Test_ProxyEmpty(t *testing.T) {
	cli, err := NewProxyClient("")
	if err != nil {
		t.Fatal(err)
	}

	ip, err := tools.GetIP(cli)
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
	cli, err := NewProxyClient("94.130.73.18:1145")
	if err != nil {
		t.Fatal(err)
	}

	ip, err := tools.GetIP(cli)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ip)
}
