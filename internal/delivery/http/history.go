package http

import (
	"fmt"
	"net/http"

	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
)

type HistoryHandler struct {
	usecase usecase.IHistoryUsecase
}

func NewHistoryHandler(h usecase.IHistoryUsecase) *HistoryHandler {
	return &HistoryHandler{
		usecase: h,
	}
}

// GetHistories godoc
// @Tags Histories
// @Summary Get histories
// @Description Get histories
// @Accept json
// @Produce json
// @Success 200 {object} domain.History
// @Router /histories [get]
// @Security     JWT
func (h *HistoryHandler) GetHistories(w http.ResponseWriter, r *http.Request) {
	var err error

	_, userId, _ := GetSession(r)
	urlParams := r.URL.Query()

	userId = userId
	urlParams = urlParams

	projectId := urlParams.Get("projectId")
	if projectId == "" {
		ErrorJSON(w, httpErrors.NewBadRequestError(fmt.Errorf("Invalid projectId")))
		return
	}

	var out struct {
		Histories []domain.History `json:"histories"`
	}
	out.Histories, err = h.usecase.Fetch()
	if err != nil {
		ErrorJSON(w, err)
		return
	}

	ResponseJSON(w, http.StatusOK, out)
}
