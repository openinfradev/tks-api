package apiClient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type ApiClient interface {
	Get(path string) (out interface{}, err error)
	Post(path string, input interface{}) (out interface{}, err error)
	Delete(path string, input interface{}) (out interface{}, err error)
}

type ApiClientImpl struct {
	client *http.Client
	url    string
	token  string
}

type ResponseJson struct {
	Code    int         `json:"status_code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// New
func New(host string, token string) (ApiClient, error) {
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
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid http status. return code: %d", res.StatusCode)
	}

	defer func() {
		res.Body.Close()
	}()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	resJson := ResponseJson{}
	if err := json.Unmarshal(body, &resJson); err != nil {
		return nil, err
	}

	if res.StatusCode != 200 || resJson.Code != 200 {
		return nil, fmt.Errorf("Invalid http status (%d). message : %s", res.StatusCode, resJson.Message)
	}

	return resJson.Data, nil
}

func (c *ApiClientImpl) Post(path string, input interface{}) (out interface{}, err error) {
	return c.callWithBody("POST", path, input)
}

func (c *ApiClientImpl) Delete(path string, input interface{}) (out interface{}, err error) {
	return c.callWithBody("DELETE", path, input)
}

func (c *ApiClientImpl) callWithBody(method string, path string, input interface{}) (out interface{}, err error) {
	pbytes, _ := json.Marshal(input)
	buff := bytes.NewBuffer(pbytes)

	req, err := http.NewRequest(method, fmt.Sprintf("%s/api/1.0/%s", c.url, path), buff)
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

	defer func() {
		res.Body.Close()
	}()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	resJson := ResponseJson{}
	if err := json.Unmarshal(body, &resJson); err != nil {
		return nil, err
	}

	if res.StatusCode != 200 || resJson.Code != 200 {
		return nil, fmt.Errorf("Invalid http status (%d). message : %s", res.StatusCode, resJson.Message)
	}

	return resJson.Data, nil

}
