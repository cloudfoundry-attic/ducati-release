package handlers

import (
	"connet-api/store"
	"lib/marshal"
	"net/http"
)

type AddRoute struct {
	Store       store.Store
	Marshaler   marshal.Marshaler
	Unmarshaler marshal.Unmarshaler
}

func (h *AddRoute) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

}
