package acceptance_test

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var (
	context     helpers.SuiteContext
	config      helpers.Config
	firstSpace  string
	secondSpace string
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)

	config = helpers.LoadConfig()

	context = helpers.NewContext(config)
	environment := helpers.NewEnvironment(context)

	BeforeSuite(func() {
		environment.Setup()

		firstSpace := context.RegularUserContext().Space
		secondSpace := firstSpace + "-2"

		cf.AsUser(context.AdminUserContext(), context.ShortTimeout(), func() {
			Eventually(cf.Cf("create-space", "-o", context.RegularUserContext().Org, secondSpace), context.ShortTimeout()).Should(gexec.Exit(0))
		})
	})

	AfterSuite(func() {
		environment.Teardown()
	})

	RunSpecs(t, "Acceptance Suite")
}

func targetSpace(spaceName string) {
	cf.AsUser(context.AdminUserContext(), context.ShortTimeout(), func() {
		Eventually(cf.Cf("target", "-s", spaceName), context.ShortTimeout()).Should(gexec.Exit(0))
	})
}
