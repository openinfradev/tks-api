package asalcm

import (
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	client *http.Client
	url    string
}

// New
func New(host string, port int, ssl bool, token string) (*Client, error) {
	var baseUrl string
	if ssl {
		if token == "" {
			return nil, fmt.Errorf("asalcm ssl enabled but token is empty.")
		}
		baseUrl = fmt.Sprintf("https://%s:%d", host, port)
	} else {
		baseUrl = fmt.Sprintf("http://%s:%d", host, port)
	}
	return &Client{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns: 10,
			},
		},
		url: baseUrl,
	}, nil
}
