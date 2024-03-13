package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	ut "github.com/go-playground/universal-translator"
	validator_ "github.com/go-playground/validator/v10"
	"github.com/openinfradev/tks-api/internal/validator"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
)

// use a single instance of Validate, it caches struct info
var (
	validate *validator_.Validate
	uni      *ut.UniversalTranslator
	trans    ut.Translator
)

func init() {
	validate, uni = validator.NewValidator()
	trans, _ = uni.GetTranslator("en")
}

func ErrorJSON(w http.ResponseWriter, r *http.Request, err error) {
	log.ErrorfWithContext(r.Context(), "error is :%s(%T)", err.Error(), err)
	errorResponse, status := httpErrors.ErrorResponse(err)
	ResponseJSON(w, r, status, errorResponse)
}

const MAX_LOG_LEN = 1000

func ResponseJSON(w http.ResponseWriter, r *http.Request, httpStatus int, data interface{}) {
	out := data

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)

	if err := json.NewEncoder(w).Encode(out); err != nil {
		log.ErrorWithContext(r.Context(), err)
	}
}

func UnmarshalRequestInput(r *http.Request, in any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	err = json.Unmarshal(body, &in)
	if err != nil {
		return err
	}

	err = validate.Struct(in)
	if err != nil {
		var valErrs validator_.ValidationErrors
		if errors.As(err, &valErrs) {
			for _, e := range valErrs {
				return httpErrors.NewBadRequestError(err, "", e.Translate(trans))
			}
		}
	}

	return nil
}

/*
func (h *APIHandler) GetClientFromClusterId(clusterId string) (*kubernetes.Clientset, error) {
	const prefix = "CACHE_KEY_KUBE_CLIENT_"
	value, found := h.Cache.Get(prefix + clusterId)
	if found {
		return value.(*kubernetes.Clientset), nil
	}
	client, err := helper.GetClientFromClusterId(clusterId)
	if err != nil {
		return nil, err
	}

	h.Cache.Set(prefix+clusterId, client, gcache.DefaultExpiration)
	return client, nil
}

func (h *APIHandler) GetKubernetesVserion() (string, error) {
	const prefix = "CACHE_KEY_KUBE_VERSION_"
	value, found := h.Cache.Get(prefix)
	if found {
		return value.(string), nil
	}
	version, err := helper.GetKubernetesVserion()
	if err != nil {
		return "", err
	}

	h.Cache.Set(prefix, version, gcache.DefaultExpiration)
	return version, nil
}

func (h *APIHandler) GetSession(r *http.Request) (string, string) {
	return r.Header.Get("ID"), r.Header.Get("AccountId")
}

func (h *APIHandler) AddHistory(r *http.Request, projectId string, historyType string, description string) error {
		userId, _ := h.GetSession(r)

		err := h.Repository.AddHistory(userId, projectId, historyType, description)
		if err != nil {
			log.ErrorWithContext(r.Context(),err)
			return err
		}

	return nil
}
*/
