package repository

// Models
type Workflow struct {
	RefID      string
	RefType    string // cluster, organization, appgroup
	WorkflowId string
	StatusDesc string
}
