package handlers

import (
	"connet-api/models"
	"connet-api/store"
	"io/ioutil"
	"lib/marshal"
	"net/http"

	"github.com/pivotal-golang/lager"
)

type AddRoute struct {
	Store       store.Store
	Logger      lager.Logger
	Unmarshaler marshal.Unmarshaler
}

func (h *AddRoute) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	logger := h.Logger.Session("add-route")

	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	var route models.Route
	err = h.Unmarshaler.Unmarshal(payload, &route)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Info("adding", lager.Data{"route": route})
	defer logger.Info("complete")

	if route.AppGuid == "" {
		logger.Error("missing-app_guid", nil)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	if route.Fqdn == "" {
		logger.Error("missing-fqdn", nil)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.Store.Create(route)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp.WriteHeader(http.StatusCreated)
	_, err = resp.Write([]byte{})
}
