package acceptance_test

import (
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const Timeout_Push = 5 * time.Minute
const Timeout_Short = 10 * time.Second

var _ = Describe("Ducati CF acceptance tests", func() {
	var proxyApp string
	var backendApp string

	BeforeEach(func() {
		proxyApp = generator.PrefixedRandomName("ducati-test-proxy-app-")
		backendApp = generator.PrefixedRandomName("ducati-test-backend-app-")

		Expect(cf.Cf("push", proxyApp, "-p", "example-apps/proxy", "-f", "example-apps/proxy/manifest.yml").Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("push", backendApp, "-p", "example-apps/proxy", "-f", "example-apps/proxy/manifest.yml").Wait(Timeout_Push)).To(gexec.Exit(0))
	})

	AfterEach(func() {
		Expect(cf.Cf("delete", proxyApp, "-f", "-r").Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("delete", backendApp, "-f", "-r").Wait(Timeout_Push)).To(gexec.Exit(0))
	})

	It("makes everything reachable", func() {

		By("checking that the proxy is reachable via its external route")
		Eventually(func() string {
			return helpers.CurlApp(proxyApp, "/")
		}, Timeout_Short).Should(ContainSubstring("hello, this is proxy"))

		By("checking that the backend is reachable via its external route")
		Eventually(func() string {
			return helpers.CurlApp(backendApp, "/")
		}, Timeout_Short).Should(ContainSubstring("hello, this is proxy"))

		By("checking that the backend is reachable via the proxy at its **external** route")
		backendWithoutScheme := backendApp + "." + helpers.LoadConfig().AppsDomain
		Eventually(func() string {
			return helpers.CurlApp(proxyApp, "/proxy/"+backendWithoutScheme)
		}, Timeout_Short).Should(ContainSubstring("hello, this is proxy"))

		// TODO: uncomment when internal DNS works
		// By("checking that the backend is reachable via the proxy at its **internal** route")
		// backendWithoutScheme = "some.internal.route.to.backend.potato"
		// Eventually(func() string {
		// 	return helpers.CurlApp(proxyApp, "/proxy/"+backendWithoutScheme)
		// }, Timeout_Short).Should(ContainSubstring("hello, this is proxy"))
	})
})
