package http

import (
	"net/http"
)

type ServerStatus = struct {
	TksContract   bool
	TksClusterLcm bool
	TksInfo       bool
	TksBatch      bool
	Database      bool
}

type ServerSettings = struct {
	Status *ServerStatus `json:"status"`
}

func (h *APIHandler) GetServerSettings(w http.ResponseWriter, r *http.Request) {
	var out struct {
		ServerSettings ServerSettings `json:"serverSettings"`
	}

	out.ServerSettings.Status = &ServerStatus{
		TksContract:   true,
		TksInfo:       true,
		TksClusterLcm: true,
		TksBatch:      true,
		Database:      true,
	}
	ResponseJSON(w, out, http.StatusOK)
}
