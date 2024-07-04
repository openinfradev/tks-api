package model

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/gorm"
)

type Stack = struct {
	gorm.Model

	ID              domain.StackId
	Name            string
	Description     string
	ClusterId       string
	OrganizationId  string
	CloudService    string
	CloudAccountId  uuid.UUID
	CloudAccount    CloudAccount
	StackTemplateId uuid.UUID
	StackTemplate   StackTemplate
	Status          domain.StackStatus
	StatusDesc      string
	PrimaryCluster  bool
	GrafanaUrl      string
	CreatorId       *uuid.UUID
	Creator         User
	UpdatorId       *uuid.UUID
	Updator         User
	Favorited       bool
	ClusterEndpoint string
	Resource        domain.DashboardStack
	PolicyIds       []string
	Conf            StackConf
	AppServeAppCnt  int
	Domains         []ClusterDomain
}

type StackConf struct {
	TksCpNode        int
	TksCpNodeMax     int
	TksCpNodeType    string
	TksInfraNode     int
	TksInfraNodeMax  int
	TksInfraNodeType string
	TksUserNode      int
	TksUserNodeMax   int
	TksUserNodeType  string
}
