package acceptance_test

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	context helpers.SuiteContext
	config  helpers.Config
)

func TestAcceptance(t *testing.T) {

	RegisterFailHandler(Fail)

	config = helpers.LoadConfig()

	context = helpers.NewContext(config)
	environment := helpers.NewEnvironment(context)

	BeforeSuite(func() {
		environment.Setup()
	})

	AfterSuite(func() {
		environment.Teardown()
	})

	RunSpecs(t, "Acceptance Suite")
}
