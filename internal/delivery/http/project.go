package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-common/pkg/log"
)

type ProjectHandler struct {
	usecase usecase.IContractUsecase
}

func NewProjectHandler(h usecase.IContractUsecase) *ProjectHandler {
	return &ProjectHandler{
		usecase: h,
	}
}

// CreateProject godoc
// @Tags Projects
// @Summary Create project
// @Description Create project
// @Accept json
// @Produce json
// @Param body body object true "body"
// @Success 200 {object} object
// @Router /projects [post]
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	userId, _ := GetSession(r)

	var input struct {
		Name        string   `json:"name"`
		Providers   []string `json:"providers"`
		GithubUrl   string   `json:"githubUrl"`
		GithubToken string   `json:"githubToken"`
		Services    []string `json:"services"`
		Description string   `json:"description"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return
	}
	err = json.Unmarshal(body, &input)
	if err != nil {
		log.Error(err)
		ErrorJSON(w, "invalid json", http.StatusBadRequest)
		return
	}

	contractId, err := h.usecase.Create(domain.Contract{
		Name:        input.Name,
		Creator:     userId,
		Description: input.Description,
	})
	if err != nil {
		log.Error("Failed to create contract err : ", err)
		//h.AddHistory(r, response.GetContractId(), "project", fmt.Sprintf("프로젝트 [%s]를 생성하는데 실패했습니다.", input.Name))
		ErrorJSON(w, fmt.Sprintf("Failed to create contract err : %s", err), http.StatusBadRequest)
		return
	}

	// add user to projectUser
	/*
		err = h.Repository.AddUserInProject(userId, response.GetContractId())
		if err != nil {
			h.AddHistory(r, response.GetContractId(), "project", fmt.Sprintf("프로젝트 [%s]를 생성하는데 실패했습니다.", input.Name))
			ErrorJSON(w, err.Error(), http.StatusInternalServerError)
			return
		}
	*/

	var out struct {
		ProjectId string `json:"id"`
	}

	out.ProjectId = contractId

	//h.AddHistory(r, response.GetContractId(), "project", fmt.Sprintf("프로젝트 [%s]를 생성하였습니다.", out.ProjectId))

	time.Sleep(time.Second * 5) // for test
	ResponseJSON(w, out, http.StatusOK)

}

// GetProjects godoc
// @Tags Projects
// @Summary Get project list
// @Description Get project list
// @Accept json
// @Produce json
// @Success 200 {object} []domain.Project
// @Router /projects [get]
func (h *ProjectHandler) GetProjects(w http.ResponseWriter, r *http.Request) {
	log.Info("GetProject")
	contracts, err := h.usecase.Fetch()
	if err != nil {
		log.Error("Failed to get contracts err : ", err)
		ErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var out struct {
		Projects []domain.Project `json:"projects"`
	}

	// Contract to Project
	for _, contract := range contracts {
		outProject := domain.Project{}
		h.reflectProject(&outProject, contract)
		out.Projects = append(out.Projects, outProject)
	}

	ResponseJSON(w, out, http.StatusOK)
}

// GetProject godoc
// @Tags Projects
// @Summary Get project detail
// @Description Get project detail
// @Accept json
// @Produce json
// @Param projectId path string true "projectId"
// @Success 200 {object} domain.Project
// @Router /projects/{projectId} [get]
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectId, ok := vars["projectId"]
	if !ok {
		ErrorJSON(w, fmt.Sprintf("Invalid input"), http.StatusBadRequest)
		return
	}

	contract, err := h.usecase.Get(projectId)
	if err != nil {
		ErrorJSON(w, fmt.Sprintf("Failed to get contract err : %s", err), http.StatusBadRequest)
		return
	}

	var out struct {
		Project domain.Project `json:"project"`
	}

	h.reflectProject(&out.Project, contract)

	ResponseJSON(w, out, http.StatusOK)
}

func (h *ProjectHandler) reflectProject(out *domain.Project, contract domain.Contract) {
	out.Id = contract.Id
	out.Name = contract.Name
	out.Description = contract.Description
	out.Status = "RUNNING"
	out.StatusDescription = ""
	out.Creator = contract.Creator
	out.CreatedAt = contract.CreatedAt
	out.UpdatedAt = contract.UpdatedAt
}
