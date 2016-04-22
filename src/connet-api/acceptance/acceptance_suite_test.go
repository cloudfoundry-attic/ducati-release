package acceptance_test

import (
	"connet-api/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"lib/testsupport"
	"math/rand"
	"net"

	. "github.com/onsi/ginkgo"
	gconfig "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

const DEFAULT_TIMEOUT = "5s"

var connetdPath string
var dbConnInfo *testsupport.DBConnectionInfo

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Connet API Acceptance Suite")
}

type beforeSuiteData struct {
	ConnetdPath string
	DBConnInfo  testsupport.DBConnectionInfo
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// only run on node 1
	fmt.Fprintf(GinkgoWriter, "building binary...")
	connetdPath, err := gexec.Build("connet-api/cmd/connetd", "-race")
	fmt.Fprintf(GinkgoWriter, "done")
	Expect(err).NotTo(HaveOccurred())

	dbConnInfo := testsupport.GetDBConnectionInfo()

	bytesToMarshal, err := json.Marshal(beforeSuiteData{
		ConnetdPath: connetdPath,
		DBConnInfo:  *dbConnInfo,
	})
	Expect(err).NotTo(HaveOccurred())

	return bytesToMarshal
}, func(marshaledBytes []byte) {
	// run on all nodes
	var data beforeSuiteData
	Expect(json.Unmarshal(marshaledBytes, &data)).To(Succeed())
	connetdPath = data.ConnetdPath
	dbConnInfo = &data.DBConnInfo

	rand.Seed(gconfig.GinkgoConfig.RandomSeed + int64(GinkgoParallelNode()))
})

var _ = SynchronizedAfterSuite(func() {
	// run on all nodes
}, func() {
	// run only on node 1
	gexec.CleanupBuildArtifacts()
})

var testDatabase *testsupport.TestDatabase

var _ = BeforeEach(func() {
	dbName := fmt.Sprintf("test_db_%x", rand.Int31())
	testDatabase = dbConnInfo.CreateDatabase(dbName)
})

var _ = AfterEach(func() {
	dbConnInfo.RemoveDatabase(testDatabase)
})

func VerifyTCPConnection(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func WriteConfigFile(connetConfig config.Config) string {
	configFile, err := ioutil.TempFile("", "test-config")
	Expect(err).NotTo(HaveOccurred())

	connetConfig.Marshal(configFile)
	Expect(configFile.Close()).To(Succeed())

	return configFile.Name()
}
