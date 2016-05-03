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
		proxyApp    string
		backendApp  string
		proxyApp2   string
		firstSpace  string
		secondSpace string
	)

	BeforeEach(func() {
		proxyApp = "proxy-app-1"
		backendApp = "backend-app"
		proxyApp2 = "proxy-app-2"

		Expect(cf.Cf("api", "--skip-ssl-validation", "api.bosh-lite.com").Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("auth", "admin", "admin").Wait(Timeout_Push)).To(gexec.Exit(0))

		firstSpace = "space1"
		Expect(cf.Cf("create-org", "test-org").Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("target", "-o", "test-org").Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("create-space", firstSpace).Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("target", "-o", "test-org", "-s", firstSpace).Wait(Timeout_Push)).To(gexec.Exit(0))

		// create binary
		// os.Setenv("GOOS", "linux")
		// os.Setenv("GOARCH", "amd64")
		// _, err := gexec.Build("example-apps/proxy")
		// Expect(err).NotTo(HaveOccurred())

		// firstSpace is currently target
		Expect(cf.Cf("push", proxyApp, "-p", "example-apps/proxy", "-f", "example-apps/proxy/manifest.yml").Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("push", backendApp, "-p", "example-apps/proxy", "-f", "example-apps/proxy/manifest.yml").Wait(Timeout_Push)).To(gexec.Exit(0))

		secondSpace = "space2"
		Expect(cf.Cf("create-space", secondSpace).Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("target", "-o", "test-org", "-s", secondSpace).Wait(Timeout_Push)).To(gexec.Exit(0))

		Expect(cf.Cf("push", proxyApp2, "-p", "example-apps/proxy", "-f", "example-apps/proxy/manifest.yml").Wait(Timeout_Push)).To(gexec.Exit(0))

		//target space1
		Expect(cf.Cf("target", "-o", "test-org", "-s", firstSpace).Wait(Timeout_Push)).To(gexec.Exit(0))
	})

	AfterEach(func() {
		Expect(cf.Cf("delete", proxyApp, "-f", "-r").Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("delete", backendApp, "-f", "-r").Wait(Timeout_Push)).To(gexec.Exit(0))

		//target space2
		Expect(cf.Cf("target", "-o", "test-org", "-s", secondSpace).Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("delete", proxyApp2, "-f", "-r").Wait(Timeout_Push)).To(gexec.Exit(0))

		Expect(cf.Cf("delete-org", "test-org", "-f").Wait(Timeout_Push)).To(gexec.Exit(0))
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

		By("checking that the backend is reachable via the proxy at its **internal** route")
		session := cf.Cf("app", backendApp, "--guid")
		Expect(session.Wait(Timeout_Push)).To(gexec.Exit(0))
		appGuid := strings.TrimSpace(string(session.Out.Contents()))
		appURL := appGuid + ".cloudfoundry"
		Eventually(func() string {
			return helpers.CurlApp(proxyApp, "/proxy/"+appURL+":8080")
		}, Timeout_Short).Should(ContainSubstring("hello, this is proxy"))

		By("checking that the backendApp is NOT reachable from a proxy app in a different space")
		Eventually(func() string {
			return helpers.CurlAppWithTimeout(proxyApp2, "/proxy/"+appURL+":8080", 60*time.Second)
		}, 4*Timeout_Short).Should(ContainSubstring("request failed"))
	})
})
