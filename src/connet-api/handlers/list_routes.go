package handlers

import (
	"connet-api/store"
	"lib/marshal"
	"net/http"

	"github.com/pivotal-golang/lager"
)

type ListRoutes struct {
	Logger    lager.Logger
	Store     store.Store
	Marshaler marshal.Marshaler
}

func (h *ListRoutes) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	logger := h.Logger.Session("list-routes")
	logger.Info("retrieving")
	defer logger.Info("complete")

	resp.Header().Set("content-type", "application/json")

	routes, err := h.Store.All()
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Info("retrieved", lager.Data{"length": len(routes)})

	routesData, err := h.Marshaler.Marshal(routes)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = resp.Write(routesData)
	if err != nil {
		logger.Error("writing-response", err)
	}
}
