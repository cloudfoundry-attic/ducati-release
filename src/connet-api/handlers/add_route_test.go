package handlers_test

import (
	"bytes"
	"connet-api/fakes"
	"connet-api/handlers"
	"connet-api/models"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/rata"
)

var _ = Describe("AddRoute", func() {
	var (
		addRoute *handlers.AddRoute
		store    *fakes.Store
		handler  http.Handler
		request  *http.Request
		payload  models.Route
	)

	BeforeEach(func() {
		store = &fakes.Store{}
		addRoute = &handlers.AddRoute{
			Store: store,
		}

		handler, request = rataWrap(addRoute, "POST", "/routes", rata.Params{})

		payload = models.Route{
			AppGuid: "my-application-guid",
			Fqdn:    "my-application-name.cloudfoundry",
		}

		payloadBytes, err := json.Marshal(payload)
		Expect(err).NotTo(HaveOccurred())
		request.Body = ioutil.NopCloser(bytes.NewBuffer(payloadBytes))
	})

	Context("when everything works", func() {
		It("succeeds with a 201 CREATED", func() {
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, request)

			Expect(resp.Code).To(Equal(http.StatusCreated))
		})

		It("creates a route in the store", func() {
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, request)

			Expect(store.CreateCallCount()).To(Equal(1))
			record := store.CreateArgsForCall(1)

			Expect(record).To(Equal(payload))
		})
	})
})
