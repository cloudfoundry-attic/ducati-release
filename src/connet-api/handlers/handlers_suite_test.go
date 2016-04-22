package handlers_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/rata"

	"testing"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handlers Suite")
}

func rataWrap(handler http.Handler, method, path string, params rata.Params) (http.Handler, *http.Request) {
	testRoutes := rata.Routes{
		{Name: "wicked_smat", Method: method, Path: path},
	}
	requestGenerator := rata.NewRequestGenerator("", testRoutes)
	testHandlers := rata.Handlers{
		"wicked_smat": handler,
	}

	router, err := rata.NewRouter(testRoutes, testHandlers)
	Expect(err).NotTo(HaveOccurred())

	request, err := requestGenerator.CreateRequest("wicked_smat", params, nil)
	Expect(err).NotTo(HaveOccurred())

	return router, request
}
