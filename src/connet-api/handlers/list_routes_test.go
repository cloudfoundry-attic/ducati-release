package handlers_test

import (
	"connet-api/fakes"
	"connet-api/handlers"
	"connet-api/models"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	lfakes "lib/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/rata"
)

var _ = Describe("ListRoutes", func() {
	var (
		logger         *lagertest.TestLogger
		handler        http.Handler
		request        *http.Request
		store          *fakes.Store
		marshaler      *lfakes.Marshaler
		expectedRoutes []models.Route
		expectedJSON   []byte
		resp           *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		marshaler = &lfakes.Marshaler{}
		marshaler.MarshalStub = json.Marshal

		expectedRoutes = []models.Route{{
			AppGuid: "my-application-guid",
			Fqdn:    "my-application-name.cloudfoundry",
		}, {
			AppGuid: "my-other-application-guid",
			Fqdn:    "my-other-application-name.cloudfoundry",
		}}

		var err error
		expectedJSON, err = json.Marshal(expectedRoutes)
		Expect(err).NotTo(HaveOccurred())

		store = &fakes.Store{}
		store.AllReturns(expectedRoutes, nil)

		logger = lagertest.NewTestLogger("test")
		listRoutesHandler := &handlers.ListRoutes{
			Logger:    logger,
			Store:     store,
			Marshaler: marshaler,
		}

		handler, request = rataWrap(listRoutesHandler, "GET", "/routes", rata.Params{})
		resp = httptest.NewRecorder()
	})

	Context("when everything works", func() {
		It("succeeds with a 200 OK", func() {
			handler.ServeHTTP(resp, request)

			Expect(resp.Code).To(Equal(http.StatusOK))
		})

		It("sets the content-type to application/json", func() {
			handler.ServeHTTP(resp, request)

			Expect(resp.Header().Get("Content-Type")).To(Equal("application/json"))
		})

		It("marshals the response from the store", func() {
			handler.ServeHTTP(resp, request)

			Expect(marshaler.MarshalCallCount()).To(Equal(1))
			Expect(marshaler.MarshalArgsForCall(0)).To(Equal(expectedRoutes))
		})

		It("should return the routes from the store", func() {
			handler.ServeHTTP(resp, request)

			Expect(store.AllCallCount()).To(Equal(1))
			Expect(resp.Body.String()).To(MatchJSON(expectedJSON))
		})
	})

	It("logs the request", func() {
		handler.ServeHTTP(resp, request)

		Expect(logger).To(gbytes.Say("list-routes.retrieving"))
		Expect(logger).To(gbytes.Say("list-routes.retrieved.*length"))
		Expect(logger).To(gbytes.Say("list-routes.complete"))
	})

	Context("when reading from the store fails", func() {
		BeforeEach(func() {
			store.AllReturns(nil, errors.New("potato"))
		})

		It("responds with a 500 status code", func() {
			handler.ServeHTTP(resp, request)

			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
		})
	})

	Context("when marshaling the response fails", func() {
		BeforeEach(func() {
			marshaler.MarshalReturns(nil, errors.New("potato"))
		})

		It("responds with a 500 status code", func() {
			handler.ServeHTTP(resp, request)

			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
		})
	})
})
