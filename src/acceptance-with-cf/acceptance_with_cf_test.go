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

const Timeout_Push = 1 * time.Minute
const Timeout_Short = 10 * time.Second

var _ = Describe("Ducati CF acceptance tests", func() {
	var appName string

	BeforeEach(func() {
		appName = generator.PrefixedRandomName("ducati-test-app-")

		Expect(cf.Cf("push", appName, "-p", "example-apps/proxy", "-f", "example-apps/proxy/manifest.yml").Wait(Timeout_Push)).To(gexec.Exit(0))
	})

	AfterEach(func() {
		Expect(cf.Cf("delete", appName, "-f", "-r").Wait(Timeout_Push)).To(gexec.Exit(0))
	})

	It("makes the app reachachable", func() {
		Eventually(func() string {
			return helpers.CurlAppRoot(appName)
		}, Timeout_Short).Should(ContainSubstring("hello, this is proxy"))
	})
})
