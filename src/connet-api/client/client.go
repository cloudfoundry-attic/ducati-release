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
	slingClient := sling.New().Client(httpClient).Base(url).Set("Accept", "application/json")

	return &connetClient{
		slingClient: slingClient,
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
	var routes []models.Route

	resp, err := c.slingClient.New().Get("/routes").Receive(&routes, nil)
	if err != nil {
		return nil, fmt.Errorf("list routes: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list routes: unexpected status code: %s", resp.Status)
	}

	return routes, nil
}
