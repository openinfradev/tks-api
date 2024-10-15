package argowf

import (
	"context"
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

func (c *ArgoClientMockImpl) GetWorkflowTemplates(ctx context.Context, namespace string) (*GetWorkflowTemplatesResponse, error) {
	return nil, nil
}

func (c *ArgoClientMockImpl) GetWorkflow(ctx context.Context, namespace string, workflowName string) (*Workflow, error) {
	return nil, nil
}

func (c *ArgoClientMockImpl) IsPausedWorkflow(ctx context.Context, namespace string, workflowName string) (bool, error) {
	return false, nil
}

func (c *ArgoClientMockImpl) GetWorkflowLog(ctx context.Context, namespace string, container string, workflowName string) (string, error) {
	return "", nil
}

func (c *ArgoClientMockImpl) GetWorkflows(ctx context.Context, namespace string) (*GetWorkflowsResponse, error) {
	return nil, nil
}

func (c *ArgoClientMockImpl) SumbitWorkflowFromWftpl(ctx context.Context, wftplName string, opts SubmitOptions) (string, error) {
	return "", nil
}

func (c *ArgoClientMockImpl) ResumeWorkflow(ctx context.Context, namespace string, workflowName string) (*Workflow, error) {
	return nil, nil
}
