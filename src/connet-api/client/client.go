package client

import (
	"connet-api/models"
	"net/http"
)

type ConnetClient struct {
}

func New(_ string, _ *http.Client) *ConnetClient {
	return nil
}

func (c *ConnetClient) AddRoute(route models.Route) error {
	panic("not implemented")
}

func (c *ConnetClient) ListRoutes() ([]models.Route, error) {
	panic("not implemented")
}
