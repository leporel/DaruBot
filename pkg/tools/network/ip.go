package network

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func GetIP(httpCli *http.Client) (string, error) {
	if httpCli == nil {
		httpCli = &http.Client{
			Timeout: 5 * time.Second,
		}
	}

	resp, err := httpCli.Get("https://api64.ipify.org/")
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cant fetch IP from api64.ipify.org: status %v", resp.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	myIP := string(bodyBytes)
	err = resp.Body.Close()
	if err != nil {
		return "", err
	}
	return myIP, nil
}
