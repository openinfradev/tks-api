package http

import (
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"net/http"
)

type IEndpointHandler interface {
	ListEndpoint(w http.ResponseWriter, r *http.Request)
}

type EndpointHandler struct {
	endpointUsecase usecase.IEndpointUsecase
}

func NewEndpointHandler(usecase usecase.Usecase) *EndpointHandler {
	return &EndpointHandler{
		endpointUsecase: usecase.Endpoint,
	}
}

// ListEndpoint godoc
//
// @Tags			Endpoint
// @Summary		List Endpoints
// @Description	List Endpoints
// @Accept			json
// @Produce		json
// @Success		200	{object}	domain.ListEndpointResponse
// @Router			/admin/endpoints [get]
// @Security		JWT
func (h EndpointHandler) ListEndpoint(w http.ResponseWriter, r *http.Request) {
	urlParams := r.URL.Query()
	pg := pagination.NewPagination(&urlParams)

	endpoints, err := h.endpointUsecase.ListEndpoints(r.Context(), pg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var out domain.ListEndpointResponse

	for _, endpoint := range endpoints {
		out.Endpoints = append(out.Endpoints, convertEndpointToDomain(endpoint))
	}

	ResponseJSON(w, r, http.StatusOK, out)
}

func convertEndpointToDomain(endpoint *model.Endpoint) domain.EndpointResponse {
	return domain.EndpointResponse{
		Name:  endpoint.Name,
		Group: endpoint.Group,
	}
}
