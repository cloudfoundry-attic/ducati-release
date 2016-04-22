package handlers_test

import (
	"bytes"
	"connet-api/fakes"
	"connet-api/handlers"
	"connet-api/models"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	lfakes "lib/fakes"
	"net/http"
	"net/http/httptest"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/rata"
)

var _ = Describe("AddRoute", func() {
	var (
		logger       *lagertest.TestLogger
		unmarshaler  *lfakes.Unmarshaler
		store        *fakes.Store
		handler      http.Handler
		request      *http.Request
		payload      models.Route
		payloadBytes []byte
		resp         *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")

		unmarshaler = &lfakes.Unmarshaler{}
		unmarshaler.UnmarshalStub = json.Unmarshal

		store = &fakes.Store{}

		addRouteHandler := &handlers.AddRoute{
			Logger:      logger,
			Store:       store,
			Unmarshaler: unmarshaler,
		}

		handler, request = rataWrap(addRouteHandler, "POST", "/routes", rata.Params{})
		resp = httptest.NewRecorder()

		payload = models.Route{
			AppGuid: "my-application-guid",
			Fqdn:    "my-application-name.cloudfoundry",
		}

		var err error
		payloadBytes, err = json.Marshal(payload)
		Expect(err).NotTo(HaveOccurred())
		request.Body = ioutil.NopCloser(bytes.NewBuffer(payloadBytes))
	})

	Context("when everything works", func() {
		It("unmarshals the payload", func() {
			handler.ServeHTTP(resp, request)

			Expect(unmarshaler.UnmarshalCallCount()).To(Equal(1))
			bytes, _ := unmarshaler.UnmarshalArgsForCall(0)
			Expect(bytes).To(Equal(payloadBytes))
		})

		It("creates a route in the store", func() {
			handler.ServeHTTP(resp, request)

			Expect(store.CreateCallCount()).To(Equal(1))
			record := store.CreateArgsForCall(0)

			Expect(record).To(Equal(payload))
		})

		It("succeeds with a 201 CREATED", func() {
			handler.ServeHTTP(resp, request)

			Expect(resp.Code).To(Equal(http.StatusCreated))
			Expect(resp.Body.String()).To(BeEmpty())
		})

		It("sets the content-type to application/json", func() {
			handler.ServeHTTP(resp, request)

			Expect(resp.Header().Get("Content-Type")).To(Equal("application/json"))
		})

		It("logs the request payload and completion", func() {
			handler.ServeHTTP(resp, request)

			Expect(logger).To(gbytes.Say("add-route.adding.*route.*my-application-guid"))
			Expect(logger).To(gbytes.Say("add-route.complete"))
		})
	})

	Context("when creating the record in the store fails", func() {
		BeforeEach(func() {
			store.CreateReturns(errors.New("cashews"))
		})

		It("fails with a 500 internal server error", func() {
			handler.ServeHTTP(resp, request)

			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
		})
	})

	Describe("payload validation", func() {
		Context("when the payload fails to unmarshal", func() {
			BeforeEach(func() {
				unmarshaler.UnmarshalReturns(errors.New("peanuts"))
			})

			It("fails with a 400 bad request", func() {
				handler.ServeHTTP(resp, request)

				Expect(resp.Code).To(Equal(http.StatusBadRequest))
			})

			It("does not create a route", func() {
				handler.ServeHTTP(resp, request)

				Expect(store.CreateCallCount()).To(Equal(0))
			})
		})

		DescribeTable("missing field responses",
			func(paramToRemove, jsonName string) {
				field := reflect.ValueOf(&payload).Elem().FieldByName(paramToRemove)
				if !field.IsValid() {
					Fail("invalid test: payload does not have a field named " + paramToRemove)
				}
				field.Set(reflect.Zero(field.Type()))

				var err error
				payloadBytes, err = json.Marshal(payload)
				Expect(err).NotTo(HaveOccurred())
				request.Body = ioutil.NopCloser(bytes.NewBuffer(payloadBytes))

				resp := httptest.NewRecorder()
				handler.ServeHTTP(resp, request)

				Expect(resp.Code).To(Equal(http.StatusBadRequest))
			},
			Entry("Application GUID", "AppGuid", "app_guid"),
			Entry("FQDN", "Fqdn", "fqdn"),
		)

		DescribeTable("missing field logging",
			func(paramToRemove, jsonName string) {
				field := reflect.ValueOf(&payload).Elem().FieldByName(paramToRemove)
				if !field.IsValid() {
					Fail("invalid test: payload does not have a field named " + paramToRemove)
				}
				field.Set(reflect.Zero(field.Type()))

				var err error
				payloadBytes, err = json.Marshal(payload)
				Expect(err).NotTo(HaveOccurred())
				request.Body = ioutil.NopCloser(bytes.NewBuffer(payloadBytes))

				resp := httptest.NewRecorder()
				handler.ServeHTTP(resp, request)

				Expect(logger).To(gbytes.Say(fmt.Sprintf("add-route.*.missing.*%s", jsonName)))
			},
			Entry("Application GUID", "AppGuid", "app_guid"),
			Entry("FQDN", "Fqdn", "fqdn"),
		)
	})
})
