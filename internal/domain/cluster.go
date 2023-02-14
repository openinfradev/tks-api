package domain

import (
	"time"
)

type Cluster = struct {
	Id                string      `json:"id"`
	ProjectId         string      `json:"projectId"`
	Name              string      `json:"name"`
	Description       string      `json:"description"`
	WorkflowId        string      `json:"workflowId"`
	Status            string      `json:"status"`
	StatusDescription string      `json:"statusDescription"`
	Conf              ClusterConf `json:"conf"`
	Creator           string      `json:"creator"`
	CreatedAt         time.Time   `json:"created"`
	UpdatedAt         time.Time   `json:"updated"`
}

type ClusterConf = struct {
	SshKeyName      string `json:"sshKeyName"`
	Region          string `json:"region"`
	MachineType     string `json:"machineType"`
	NumOfAz         int    `json:"numOfAz"`
	MinSizePerAz    int    `json:"minSizePerAz"`
	MaxSizePerAz    int    `json:"maxSizePerAz"`
	MachineReplicas int    `json:"machineReplicas"`
}

type ClusterCapacity = struct {
	Max     int
	Current int
}

type ClusterKubeInfo = struct {
	Version        string          `json:"version"`
	TotalResources int             `json:"totalResources"`
	Nodes          int             `json:"nodes"`
	Namespaces     int             `json:"namespaces"`
	Services       int             `json:"services"`
	Pods           int             `json:"pods"`
	Cores          ClusterCapacity `json:"cores"`
	Memory         ClusterCapacity `json:"memory"`
	Updated        time.Time       `json:"updated"`
}

type Event = struct {
	Id        string    `json:"id"`
	Namespace string    `json:"namespace"`
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Updated   time.Time `json:"updated"`
}

type Node = struct {
	Id           string    `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	InstanceType string    `json:"instanceType"`
	Role         string    `json:"role"`
	Updated      time.Time `json:"updated"`
}
