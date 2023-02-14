package http

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

type InfraProvider = struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	AccessId    string    `json:"accessId"`
	AccessKey   string    `json:"accessKey"`
	Creator     string    `json:"creator"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

func (h *APIHandler) GetInfraProviders(w http.ResponseWriter, r *http.Request) {
	var out struct {
		InfraProviders []InfraProvider `json:"infraProviders"`
	}
	out.InfraProviders = make([]InfraProvider, 2)
	out.InfraProviders[0].Id = uuid.New().String()
	out.InfraProviders[0].Name = "default"
	out.InfraProviders[0].Type = "AWS"
	out.InfraProviders[0].Description = "system default"
	out.InfraProviders[0].AccessId = "ACCESSID_ASDF"
	out.InfraProviders[0].AccessKey = "ACCESSKEY_SDAFSF"
	out.InfraProviders[0].Creator = "taekyu.kang"
	out.InfraProviders[0].Created = time.Now()
	out.InfraProviders[0].Updated = time.Now()

	out.InfraProviders[1].Id = uuid.New().String()
	out.InfraProviders[1].Name = "default"
	out.InfraProviders[1].Type = "BYOH"
	out.InfraProviders[1].Description = "system default"
	out.InfraProviders[1].AccessId = ""
	out.InfraProviders[1].AccessKey = ""
	out.InfraProviders[1].Creator = "taekyu.kang"
	out.InfraProviders[1].Created = time.Now()
	out.InfraProviders[1].Updated = time.Now()

	ResponseJSON(w, out, http.StatusOK)
}

func (h *APIHandler) GetInfraProvider(w http.ResponseWriter, r *http.Request) {
	var out struct {
		InfraProvider InfraProvider `json:"infraProvider"`
	}

	out.InfraProvider.Id = uuid.New().String()
	out.InfraProvider.Name = "AWS"
	out.InfraProvider.Type = "AWS"
	out.InfraProvider.Description = "infra provider - AWS"
	out.InfraProvider.AccessId = "ACCESSID_ASDF"
	out.InfraProvider.AccessKey = "ACCESSKEY_SDAFSF"
	out.InfraProvider.Creator = "taekyu.kang"
	out.InfraProvider.Created = time.Now()
	out.InfraProvider.Updated = time.Now()

	ResponseJSON(w, out, http.StatusOK)
}

func (h *APIHandler) CreateInfraProvider(w http.ResponseWriter, r *http.Request) {
	var out struct {
		InfraProvider InfraProvider `json:"infraProvider"`
	}

	ResponseJSON(w, out, http.StatusOK)
}
