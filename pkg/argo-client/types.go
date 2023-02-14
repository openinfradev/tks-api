package argowf

// GetWorkflowTemplatesResponse is a response from GET /api/v1/workflow-templates API.
type GetWorkflowTemplatesResponse struct {
	Items []WorkflowTemplate `json:"items"`
}

type WorkflowTemplate struct {
	Metadata Metadata             `json:"metadata"`
	Spec     WorkflowTemplateSpec `json:"spec"`
}

type WorkflowTemplateSpec struct {
	Args Arguments `json:"arguments"`
}

type Metadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Arguments struct {
	Parameters []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"parameters"`
}

type submitWorkflowRequestBody struct {
	Namespace     string        `json:"namespace,omitempty"`
	ResourceKind  string        `json:"resourceKind,omitempty"`
	ResourceName  string        `json:"resourceName,omitempty"`
	SubmitOptions SubmitOptions `json:"submitOptions,omitempty"`
}

// SubmitOptions is optional fields to submit new workflow.
type SubmitOptions struct {
	DryRun       bool     `json:"dryRun,omitempty"`
	EntryPoint   string   `json:"entryPoint,omitempty"`
	GenerateName string   `json:"generateName,omitempty"`
	Labels       string   `json:"labels,omitempty"`
	Name         string   `json:"name,omitempty"`
	Parameters   []string `json:"parameters,omitempty"`
}

type SubmitWorkflowResponse struct {
	Metadata Metadata `json:"metadata"`
}

type GetWorkflowsResponse struct {
	Items []Workflow `json:"items"`
}

type Workflow struct {
	ApiVersion string           `json:"apiVersion"`
	Metadata   WorkflowMetadata `json:"metadata"`
	Spec       WorkflowSpec     `json:"spec"`
	Status     WorkflowStatus   `json:"status"`
}

type WorkflowMetadata struct {
	ClusterName  string `json:"clusterName"`
	GenerateName string `json:"generateName"`
	Name         string `json:"name"`
	NameSpace    string `json:"namespace"`
}

type WorkflowSpec struct {
	Args WorkflowArguments `json:"arguments"`
}

type WorkflowArguments struct {
	Parameters []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"parameters"`
}

type WorkflowStatus struct {
	Phase    string `json:"phase"`
	Progress string `json:"progress"`
	Message  string `json:"message"`
}
