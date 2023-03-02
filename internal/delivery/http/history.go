package http

import (
	"net/http"

	"github.com/openinfradev/tks-api/internal/domain"
	"github.com/openinfradev/tks-api/internal/usecase"
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

	userId, _ := GetSession(r)
	urlParams := r.URL.Query()

	userId = userId
	urlParams = urlParams

	projectId := urlParams.Get("projectId")
	if projectId == "" {
		ErrorJSON(w, "Invalid projectId", http.StatusOK)
		return
	}

	var out struct {
		Histories []domain.History `json:"histories"`
	}
	out.Histories, err = h.usecase.Fetch()
	if err != nil {
		ErrorJSON(w, "failed to fetch histories", http.StatusOK)
		return
	}

	ResponseJSON(w, out, http.StatusOK)
}
