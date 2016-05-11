package acceptance_test

import (
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const Timeout_Push = 5 * time.Minute
const Timeout_Short = 10 * time.Second

var _ = Describe("Ducati CF acceptance tests", func() {
	var (
		proxyApp      string
		backendApp    string
		backendAppURL string
		proxyApp2     string
		orgName       string
	)

	BeforeEach(func() {
		proxyApp = "proxy-app-1"
		backendApp = "backend-app"
		proxyApp2 = "proxy-app-2"

		orgName = "test-org"
		Expect(cf.Cf("create-org", orgName).Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("target", "-o", orgName).Wait(Timeout_Push)).To(gexec.Exit(0))

		firstSpace := "space1"
		Expect(cf.Cf("create-space", firstSpace).Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("target", "-o", orgName, "-s", firstSpace).Wait(Timeout_Push)).To(gexec.Exit(0))

		pushApp(proxyApp)
		pushApp(backendApp)

		session := cf.Cf("app", backendApp, "--guid")
		Expect(session.Wait(Timeout_Push)).To(gexec.Exit(0))
		backendAppGuid := strings.TrimSpace(string(session.Out.Contents()))
		backendAppURL = backendAppGuid + ".cloudfoundry"

		secondSpace := "space2"
		Expect(cf.Cf("create-space", secondSpace).Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("target", "-o", orgName, "-s", secondSpace).Wait(Timeout_Push)).To(gexec.Exit(0))

		pushApp(proxyApp2)
	})

	AfterEach(func() {
		// clean up everything
		Expect(cf.Cf("delete-org", orgName, "-f").Wait(Timeout_Push)).To(gexec.Exit(0))
	})

	It("makes everything reachable", func() {
		By("checking that the proxy is reachable via its external route")
		Eventually(func() string {
			return helpers.CurlAppWithTimeout(proxyApp, "/", 6*Timeout_Short)
		}, 6*Timeout_Short, time.Second).Should(ContainSubstring("hello, this is proxy"))

		By("checking that the backend is reachable via its external route")
		Eventually(func() string {
			return helpers.CurlAppWithTimeout(backendApp, "/", 6*Timeout_Short)
		}, 6*Timeout_Short, time.Second).Should(ContainSubstring("hello, this is proxy"))

		By("checking that the backend is reachable via the proxy at its **external** route")
		backendWithoutScheme := backendApp + "." + helpers.LoadConfig().AppsDomain
		Eventually(func() string {
			return helpers.CurlAppWithTimeout(proxyApp, "/proxy/"+backendWithoutScheme, 6*Timeout_Short)
		}, 6*Timeout_Short, time.Second).Should(ContainSubstring("hello, this is proxy"))

		By("checking that the backend is reachable via the proxy at its **internal** route")
		Eventually(func() string {
			return helpers.CurlAppWithTimeout(proxyApp, "/proxy/"+backendAppURL+":8080", 6*Timeout_Short)
		}, 6*Timeout_Short, time.Second).Should(ContainSubstring("hello, this is proxy"))

		By("checking that the backendApp is NOT reachable from a proxy app in a different space")
		Eventually(func() string {
			return helpers.CurlAppWithTimeout(proxyApp2, "/proxy/"+backendAppURL+":8080", 6*Timeout_Short)
		}, 6*Timeout_Short, time.Second).Should(ContainSubstring("request failed"))
	})
})
