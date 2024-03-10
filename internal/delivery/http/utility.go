package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

type UtilityHandler struct {
	usecase usecase.IUtilityUsecase
}

type IUtilityHandler interface {
	RegoCompile(w http.ResponseWriter, r *http.Request)
}

func NewUtilityHandler(u usecase.Usecase) IUtilityHandler {
	return &UtilityHandler{
		usecase: u.Utility,
	}
}

// CompileRego godoc
//
//	@Tags			Rego
//	@Summary		[CompileRego] Rego 코드 컴파일 및 파라미터 파싱
//	@Description	Rego 코드 컴파일 및 파라미터 파싱을 수행한다. 파라미터 파싱을 위해서는 먼저 컴파일이 성공해야 하며, parseParameter를 false로 하면 컴파일만 수행할 수 있다.
//	@Accept			json
//	@Produce		json
//	@Param			parseParameter	query		bool						true	"파라미터 파싱 여부"
//	@Param			body			body		domain.RegoCompileRequest	true	"Rego 코드"
//	@Success		200				{object}	domain.RegoCompileResponse
//	@Router			/utility/rego-compile [post]
//	@Security		JWT
func (h *UtilityHandler) RegoCompile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parseParameter := false

	parse, ok := vars["parseParameter"]
	if ok {
		parsedBool, err := strconv.ParseBool(parse)
		if err != nil {
			ErrorJSON(w, r, httpErrors.NewBadRequestError(fmt.Errorf("invalid parseParameter: '%s'", parse), "U_INVALID_REGO_PARSEPARAM", ""))
			return
		}
		parseParameter = parsedBool
	}

	input := domain.RegoCompileRequest{}
	err := UnmarshalRequestInput(r, &input)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	response, err := h.usecase.RegoCompile(&input, parseParameter)
	if err != nil {
		log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)

		ErrorJSON(w, r, err)
		return
	}

	ResponseJSON(w, r, http.StatusCreated, response)
}
