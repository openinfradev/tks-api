package http

import (
	"github.com/openinfradev/tks-api/internal/usecase"
	"net/http"
)

type IProjectHandler interface {
	CreateProject(w http.ResponseWriter, r *http.Request)
	UpdateProject(w http.ResponseWriter, r *http.Request)
	DeleteProject(w http.ResponseWriter, r *http.Request)
	GetProject(w http.ResponseWriter, r *http.Request)
	GetProjects(w http.ResponseWriter, r *http.Request)

	AddProjectMember(w http.ResponseWriter, r *http.Request)
	RemoveProjectMember(w http.ResponseWriter, r *http.Request)
	GetProjectMembers(w http.ResponseWriter, r *http.Request)
	UpdateProjectMemberRole(w http.ResponseWriter, r *http.Request)

	CreateProjectNamespace(w http.ResponseWriter, r *http.Request)
	GetProjectNamespaces(w http.ResponseWriter, r *http.Request)
	GetProjectNamespace(w http.ResponseWriter, r *http.Request)
	DeleteProjectNamespace(w http.ResponseWriter, r *http.Request)
}
type ProjectHandler struct {
	usecase usecase.IProjectUsecase
}

func (p ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (p ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (p ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (p ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (p ProjectHandler) GetProjects(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (p ProjectHandler) AddProjectMember(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (p ProjectHandler) RemoveProjectMember(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (p ProjectHandler) GetProjectMembers(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (p ProjectHandler) UpdateProjectMemberRole(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (p ProjectHandler) CreateProjectNamespace(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (p ProjectHandler) GetProjectNamespaces(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (p ProjectHandler) GetProjectNamespace(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (p ProjectHandler) DeleteProjectNamespace(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func NewProjectHandler(u usecase.IProjectUsecase) IProjectHandler {
	return &ProjectHandler{
		usecase: u,
	}
}
