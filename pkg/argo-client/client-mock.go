package argowf

import (
	"net/http"
	"time"
)

type ArgoClientMockImpl struct {
	client *http.Client
	url    string
}

// New
func NewMock() (ArgoClient, error) {
	return &ArgoClientMockImpl{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns: 10,
			},
		},
		url: "",
	}, nil
}

func (c *ArgoClientMockImpl) GetWorkflowTemplates(namespace string) (*GetWorkflowTemplatesResponse, error) {
	return nil, nil
}

func (c *ArgoClientMockImpl) GetWorkflow(namespace string, workflowName string) (*Workflow, error) {
	return nil, nil
}

func (c *ArgoClientMockImpl) GetWorkflows(namespace string) (*GetWorkflowsResponse, error) {
	return nil, nil
}

func (c *ArgoClientMockImpl) SumbitWorkflowFromWftpl(wftplName string, opts SubmitOptions) (string, error) {
	return "", nil
}
