package client_test

import (
	"connet-api/client"
	"connet-api/models"
	"errors"
	"net/http"

	"github.com/cloudfoundry-incubator/ducati-daemon/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Client", func() {
	var (
		roundTripper *fakes.RoundTripper
		server       *ghttp.Server

		c client.ConnetClient
	)

	BeforeEach(func() {
		roundTripper = &fakes.RoundTripper{}
		roundTripper.RoundTripStub = http.DefaultTransport.RoundTrip

		server = ghttp.NewServer()

		httpClient := &http.Client{
			Transport: roundTripper,
		}

		c = client.New(server.URL(), httpClient)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Create", func() {
		var route models.Route

		BeforeEach(func() {
			route = models.Route{
				AppGuid: "my-application-guid",
				Fqdn:    "my-application-name.cloudfoundry",
			}

			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/routes"),
				ghttp.VerifyJSONRepresenting(route),
				ghttp.VerifyHeaderKV("Content-type", "application/json"),
				ghttp.RespondWithJSONEncoded(http.StatusCreated, ""),
			))
		})

		It("POSTs a route payload to /routes", func() {
			err := c.AddRoute(route)
			Expect(err).NotTo(HaveOccurred())

			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		It("uses the provided http client", func() {
			err := c.AddRoute(route)
			Expect(err).NotTo(HaveOccurred())

			Expect(roundTripper.RoundTripCallCount()).To(Equal(1))
		})

		Context("when the request fails", func() {
			BeforeEach(func() {
				roundTripper.RoundTripReturns(nil, errors.New("potato"))
			})

			It("returns the error", func() {
				err := c.AddRoute(route)
				Expect(err).To(MatchError(MatchRegexp("add route:.*potato")))
			})
		})

		Context("when the response is not 201 StatusCreated", func() {
			BeforeEach(func() {
				server.RouteToHandler("POST", "/routes", ghttp.CombineHandlers(
					ghttp.RespondWithJSONEncoded(http.StatusBadRequest, ""),
				))
			})

			It("returns an error", func() {
				err := c.AddRoute(route)

				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).To(MatchError("add route: unexpected status code: 400 Bad Request"))
			})
		})
	})
})
