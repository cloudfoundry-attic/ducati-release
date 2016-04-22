package client

import (
	"connet-api/models"
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
)

type connetClient struct {
	slingClient *sling.Sling
}

type ConnetClient interface {
	AddRoute(models.Route) error
	ListRoutes() ([]models.Route, error)
}

func New(url string, httpClient *http.Client) ConnetClient {
	return &connetClient{
		slingClient: sling.New().Client(httpClient).Base(url),
	}
}

func (c *connetClient) AddRoute(route models.Route) error {
	resp, err := c.slingClient.New().Post("/routes").BodyJSON(route).Receive(nil, nil)
	if err != nil {
		return fmt.Errorf("add route: %s", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("add route: unexpected status code: %s", resp.Status)
	}

	return nil
}

func (c *connetClient) ListRoutes() ([]models.Route, error) {
	panic("not implemented")
}
