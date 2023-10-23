package apiClient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

type ApiClient interface {
	Get(path string) (out interface{}, err error)
	Post(path string, input interface{}) (out interface{}, err error)
	Delete(path string, input interface{}) (out interface{}, err error)
	Put(path string, input interface{}) (out interface{}, err error)
	Patch(path string, input interface{}) (out interface{}, err error)
	SetToken(token string)
}

type ApiClientImpl struct {
	client *http.Client
	url    string
	token  string
}

// New
func NewWithToken(host string, token string) (ApiClient, error) {
	return &ApiClientImpl{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns: 10,
			},
		},
		url:   host,
		token: token,
	}, nil
}

func New(host string) (ApiClient, error) {
	return &ApiClientImpl{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns: 10,
			},
		},
		url:   host,
		token: "",
	}, nil
}

func (c *ApiClientImpl) SetToken(token string) {
	c.token = token
}

func (c *ApiClientImpl) Get(path string) (out interface{}, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/1.0/%s", c.url, path), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.token)
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("Failed to call api server.")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		res.Body.Close()
	}()

	if res.StatusCode != 200 {
		var restError httpErrors.RestError
		if err := json.Unmarshal(body, &restError); err != nil {
			return nil, fmt.Errorf("Invalid http status. failed to unmarshal body : %s", err)
		}

		return restError, fmt.Errorf("HTTP status [%d] message [%s]", res.StatusCode, restError.ErrMessage)
	}

	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func (c *ApiClientImpl) Post(path string, input interface{}) (out interface{}, err error) {
	return c.callWithBody("api", "POST", path, input)
}

func (c *ApiClientImpl) Delete(path string, input interface{}) (out interface{}, err error) {
	return c.callWithBody("api", "DELETE", path, input)
}

func (c *ApiClientImpl) Put(path string, input interface{}) (out interface{}, err error) {
	return c.callWithBody("api", "PUT", path, input)
}

func (c *ApiClientImpl) Patch(path string, input interface{}) (out interface{}, err error) {
	return c.callWithBody("api", "PATCH", path, input)
}

func (c *ApiClientImpl) callWithBody(prefix string, method string, path string, input interface{}) (out interface{}, err error) {
	pbytes, _ := json.Marshal(input)
	buff := bytes.NewBuffer(pbytes)

	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s/1.0/%s", c.url, prefix, path), buff)
	if err != nil {
		return nil, err
	}

	if prefix == "api" {
		req.Header.Add("Authorization", "Bearer "+c.token)
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("Failed to call api server.")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		res.Body.Close()
	}()

	if res.StatusCode != 200 {
		var restError httpErrors.RestError
		if err := json.Unmarshal(body, &restError); err != nil {
			return nil, fmt.Errorf("Invalid http status. failed to unmarshal body : %s", err)
		}
		return restError, fmt.Errorf("HTTP status [%d] message [%s]", res.StatusCode, restError.ErrMessage)
	}

	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}

	return
}
