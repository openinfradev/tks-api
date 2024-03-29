package thanos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/pkg/log"
)

type ThanosClient interface {
	Get(ctx context.Context, query string) (Metric, error)
	FetchRange(ctx context.Context, query string, start int, end int, step int) (out Metric, err error)
}

type ThanosClientImpl struct {
	client *http.Client
	url    string
}

// New
func New(host string, port int, ssl bool, token string) (ThanosClient, error) {
	var baseUrl string
	if ssl {
		if token == "" {
			return nil, fmt.Errorf("thanos ssl enabled but token is empty.")
		}
		baseUrl = fmt.Sprintf("%s:%d", host, port)
	} else {
		baseUrl = fmt.Sprintf("%s:%d", host, port)
	}
	return &ThanosClientImpl{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns: 10,
			},
		},
		url: baseUrl,
	}, nil
}

func (c *ThanosClientImpl) Get(ctx context.Context, query string) (out Metric, err error) {
	url := c.url + "/api/v1/query?query=" + url.QueryEscape(query)

	log.Info(ctx, "url : ", url)
	res, err := c.client.Get(url)
	if err != nil {
		return out, err
	}
	if res == nil {
		return out, fmt.Errorf("Failed to call thanos.")
	}
	if res.StatusCode != 200 {
		return out, fmt.Errorf("Invalid http status. return code: %d", res.StatusCode)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Error(ctx, "error closing http body")
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	err = json.Unmarshal(body, &out)
	if err != nil {
		return out, err
	}

	log.Info(ctx, helper.ModelToJson(out))
	return
}

func (c *ThanosClientImpl) FetchRange(ctx context.Context, query string, start int, end int, step int) (out Metric, err error) {
	rangeParam := fmt.Sprintf("&dedup=true&partial_response=false&start=%d&end=%d&step=%d&max_source_resolution=0s", start, end, step)
	query = url.QueryEscape(query) + rangeParam
	url := c.url + "/api/v1/query_range?query=" + query

	log.Info(ctx, "url : ", url)
	res, err := c.client.Get(url)
	if err != nil {
		return out, err
	}
	if res == nil {
		return out, fmt.Errorf("Failed to call thanos.")
	}
	if res.StatusCode != 200 {
		return out, fmt.Errorf("Invalid http status. return code: %d", res.StatusCode)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Error(ctx, "error closing http body")
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	/*
		var a interface{}
		err = json.Unmarshal(body, &a)
		if err != nil {
			return out, err
		}
		log.Info(helper.ModelToJson(a))
	*/

	err = json.Unmarshal(body, &out)
	if err != nil {
		return out, err
	}

	return
}
