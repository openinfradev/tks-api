package argowf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/openinfradev/tks-api/pkg/log"
)

type ArgoClient interface {
	GetWorkflowTemplates(namespace string) (*GetWorkflowTemplatesResponse, error)
	GetWorkflow(namespace string, workflowName string) (*Workflow, error)
	GetWorkflows(namespace string) (*GetWorkflowsResponse, error)
	SumbitWorkflowFromWftpl(wftplName string, opts SubmitOptions) (string, error)
}

type ArgoClientImpl struct {
	client *http.Client
	url    string
}

// New
func New(host string, port int, ssl bool, token string) (ArgoClient, error) {
	var baseUrl string
	if ssl {
		if token == "" {
			return nil, fmt.Errorf("argo ssl enabled but token is empty.")
		}
		baseUrl = fmt.Sprintf("%s:%d", host, port)
	} else {
		baseUrl = fmt.Sprintf("%s:%d", host, port)
	}
	return &ArgoClientImpl{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns: 10,
			},
		},
		url: baseUrl,
	}, nil
}

func (c *ArgoClientImpl) GetWorkflowTemplates(namespace string) (*GetWorkflowTemplatesResponse, error) {
	res, err := http.Get(fmt.Sprintf("%s/api/v1/workflow-templates/%s", c.url, namespace))
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("Failed to call argo workflow.")
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid http status. return code: %d", res.StatusCode)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Error("error closing http body")
		}
	}()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	wftplRes := GetWorkflowTemplatesResponse{}
	if err := json.Unmarshal(body, &wftplRes); err != nil {
		log.Error("an error was unexpected while parsing response from api /workflow template.")
		return nil, err
	}
	return &wftplRes, nil
}

func (c *ArgoClientImpl) GetWorkflow(namespace string, workflowName string) (*Workflow, error) {
	res, err := http.Get(fmt.Sprintf("%s/api/v1/workflows/%s/%s", c.url, namespace, workflowName))
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("Failed to call argo workflow.")
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid http status. return code: %d", res.StatusCode)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Error("error closing http body")
		}
	}()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	workflowRes := Workflow{}
	if err := json.Unmarshal(body, &workflowRes); err != nil {
		log.Error("an error was unexpected while parsing response from api /workflow template.")
		return nil, err
	}

	return &workflowRes, nil
}

func (c *ArgoClientImpl) GetWorkflows(namespace string) (*GetWorkflowsResponse, error) {
	res, err := http.Get(fmt.Sprintf("%s/api/v1/workflows/%s", c.url, namespace))
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("Failed to call argo workflow.")
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid http status. return code: %d", res.StatusCode)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Error("error closing http body")
		}
	}()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	workflowsRes := GetWorkflowsResponse{}
	if err := json.Unmarshal(body, &workflowsRes); err != nil {
		log.Error("an error was unexpected while parsing response from api /workflow template.")
		return nil, err
	}

	return &workflowsRes, nil
}

func (c *ArgoClientImpl) SumbitWorkflowFromWftpl(wftplName string, opts SubmitOptions) (string, error) {
	reqBody := submitWorkflowRequestBody{
		Namespace:     "argo",
		ResourceKind:  "WorkflowTemplate",
		ResourceName:  wftplName,
		SubmitOptions: opts,
	}
	log.Debug("SumbitWorkflowFromWftpl reqBody ", reqBody)

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "",
			fmt.Errorf("an error was unexpected while marshaling request body")
	}
	buff := bytes.NewBuffer(reqBodyBytes)

	res, err := http.Post(fmt.Sprintf("%s/api/v1/workflows/argo/submit", c.url), "application/json", buff)
	if err != nil {
		return "", err
	}
	if res == nil {
		return "", fmt.Errorf("Failed to call argo workflow.")
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("Invalid http status. return code: %d", res.StatusCode)
	}

	defer func() {
		if res != nil {
			if err := res.Body.Close(); err != nil {
				log.Error("error closing http body")
			}
		}
	}()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	submitRes := SubmitWorkflowResponse{}
	if err := json.Unmarshal(body, &submitRes); err != nil {
		log.Error("an error was unexpected while parsing response from api /submit.")
		return "", err
	}
	return submitRes.Metadata.Name, nil
}
