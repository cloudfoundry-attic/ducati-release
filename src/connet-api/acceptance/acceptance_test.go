package acceptance_test

import (
	"connet-api/client"
	"connet-api/config"
	"connet-api/models"
	"fmt"
	"net/http"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Acceptance", func() {
	var (
		session      *gexec.Session
		conf         config.Config
		connetClient client.ConnetClient
		address      string
	)

	var serverIsAvailable = func() error {
		return VerifyTCPConnection(address)
	}

	BeforeEach(func() {
		conf = config.Config{
			ListenHost: "127.0.0.1",
			ListenPort: 9001 + GinkgoParallelNode(),
			Database:   testDatabase.DBConfig(),
		}
		configFilePath := WriteConfigFile(conf)

		connetdCmd := exec.Command(connetdPath, "-configFile", configFilePath)
		var err error
		session, err = gexec.Start(connetdCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		address = fmt.Sprintf("%s:%d", conf.ListenHost, conf.ListenPort)
		connetClient = client.New("http://"+address, http.DefaultClient)

		Eventually(serverIsAvailable, DEFAULT_TIMEOUT).Should(Succeed())
	})

	AfterEach(func() {
		session.Interrupt()
		Eventually(session, DEFAULT_TIMEOUT).Should(gexec.Exit(0))
	})

	It("should boot and gracefully terminate", func() {
		Consistently(session).ShouldNot(gexec.Exit())

		session.Interrupt()
		Eventually(session, DEFAULT_TIMEOUT).Should(gexec.Exit(0))
	})

	Describe("adding and listing routes", func() {
		It("supports adding a route", func() {
			newRoute := models.Route{
				AppGuid: "some-app-guid",
				Fqdn:    "some.fully.qualified.domain",
			}
			err := connetClient.AddRoute(newRoute)
			Expect(err).NotTo(HaveOccurred())

			routes, err := connetClient.ListRoutes()
			Expect(err).NotTo(HaveOccurred())

			Expect(routes).To(ContainElement(newRoute))
		})

	})
})
