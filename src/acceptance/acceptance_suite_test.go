package acceptance_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Suite")
}

var gardenServer1, gardenServer2 string
var pathToDucatiDNSBinary string

var _ = BeforeSuite(func() {
	gardenServer1 = os.Getenv("GARDEN_SERVER_1")
	if gardenServer1 == "" {
		Fail("missing required env var GARDEN_SERVER_1")
	}
	gardenServer2 = os.Getenv("GARDEN_SERVER_2")
	if gardenServer2 == "" {
		Fail("missing required env var GARDEN_SERVER_2")
	}

	var err error
	pathToDucatiDNSBinary, err = gexec.Build("github.com/cloudfoundry-incubator/ducati-dns/cmd/ducati-dns")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
