package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gotest.tools/v3/assert"
)

var (
	createdStackTemplateId uuid.UUID
)

func init() {
}

// Tests
func TestCreateStackTemplate(t *testing.T) {
	body := domain.CreateStackTemplateRequest{
		Name:            "testName",
		Description:     "testDescription",
		Version:         "v1",
		CloudService:    "AWS",
		Platform:        "AWS",
		TemplateType:    "STANDARD",
		Template:        "aws-standard",
		KubeVersion:     "v1.21",
		KubeType:        "AWS",
		OrganizationIds: []string{"master"},
		ServiceIds:      []string{"LMA", "SERVICE_MESH"},
	}

	bodyBytes, _ := json.Marshal(body)
	r, _ := http.NewRequest("POST", "/admin/stack-templates", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()

	vars := map[string]string{
		"mystring": "abcd",
	}
	r = mux.SetURLVars(r, vars)

	rf := repository.NewRepositoryFactory(db)
	uf := usecase.NewUsecaseFactory(*rf)
	handler := NewStackTemplateHandler(*uf)
	handler.CreateStackTemplate(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []byte("abcd"), w.Body.Bytes())
}
