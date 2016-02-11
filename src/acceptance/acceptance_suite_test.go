package acceptance_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Suite")
}

var gardenServer1, gardenServer2 string

var _ = BeforeSuite(func() {
	gardenServer1 = os.Getenv("GARDEN_SERVER_1")
	if gardenServer1 == "" {
		Fail("missing required env var GARDEN_SERVER_1")
	}
	gardenServer2 = os.Getenv("GARDEN_SERVER_2")
	if gardenServer2 == "" {
		Fail("missing required env var GARDEN_SERVER_2")
	}
})
