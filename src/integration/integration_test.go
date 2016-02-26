package integration_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"

	"github.com/cloudfoundry-incubator/ducati-daemon/lib/namespace"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	testsupport "github.com/cloudfoundry-incubator/ducati-daemon/testsupport"
)

const DEFAULT_TIMEOUT = "5s"

var _ = Describe("how the VXLAN plugin talks to the ducati daemon", func() {
	var (
		daemonSession  *gexec.Session
		pluginSession  *gexec.Session
		address        string
		subnet         string
		overlay        string
		repoDir        string
		containerNS    namespace.Namespace
		containerID    string
		netConfig      Config
		serverURL      string
		testDatabase   *testsupport.TestDatabase
		sandboxRepoDir string
	)

	BeforeEach(func() {
		dbName := fmt.Sprintf("test_db_%x", rand.Int31())
		testDatabase = dbConnInfo.CreateDatabase(dbName)

		By("booting the daemon")
		address = fmt.Sprintf("127.0.0.1:%d", 4001+GinkgoParallelNode())
		serverURL = "http://" + address
		subnet = "192.168.1.1/24"
		overlay = "192.168.0.0/16"

		var err error
		sandboxRepoDir, err = ioutil.TempDir("", "sandbox")
		Expect(err).NotTo(HaveOccurred())

		daemonCmd := exec.Command(pathToDaemon,
			"-listenAddr", address,
			"-overlayNetwork", overlay,
			"-localSubnet", subnet,
			"-databaseURL", testDatabase.URL(),
			"-sandboxRepoDir", sandboxRepoDir,
		)

		daemonSession, err = gexec.Start(daemonCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		By("creating a container")
		containerID = fmt.Sprintf("%x", rand.Int31())
		repoDir, err = ioutil.TempDir("", "namespaces-")
		Expect(err).NotTo(HaveOccurred())

		namespaceRepo, err := namespace.NewRepository(repoDir)
		Expect(err).NotTo(HaveOccurred())

		containerNS, err = namespaceRepo.Create("container-ns")
		Expect(err).NotTo(HaveOccurred())

		netConfig = Config{
			Type:      "vxlan",
			NetworkID: "network-id",
		}
	})

	AfterEach(func() {
		Expect(containerNS.Destroy()).To(Succeed())
		Expect(os.RemoveAll(repoDir)).To(Succeed())

		daemonSession.Interrupt()
		Eventually(daemonSession, DEFAULT_TIMEOUT).Should(gexec.Exit(0))
		dbConnInfo.RemoveDatabase(testDatabase)
	})

	var serverIsAvailable = func() error {
		_, err := net.Dial("tcp", address)
		return err
	}

	It("should maintain the container state in the daemon", func() {
		url := fmt.Sprintf("http://%s/containers", address)
		Eventually(serverIsAvailable, DEFAULT_TIMEOUT).Should(Succeed())

		By("checking that the daemon knows of no containers")
		resp, err := http.Get(url)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()

		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		respBytes, err := ioutil.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())

		Expect(respBytes).To(MatchJSON(`[]`))

		By("invoking the vxlan CNI plugin with the ADD action")
		addCmd, err := buildCNICmd("ADD", netConfig, containerNS, containerID, repoDir, serverURL)
		Expect(err).NotTo(HaveOccurred())
		pluginSession, err = gexec.Start(addCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(pluginSession, DEFAULT_TIMEOUT).Should(gexec.Exit(0))

		By("checking that the daemon now has the container data")
		resp, err = http.Get(url)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()

		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		respBytes, err = ioutil.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())

		type containerData struct {
			ID     string `json:"id"`
			IP     string `json:"ip"`
			MAC    string `json:"mac"`
			HostIP string `json:"host_ip"`
		}

		var output []*containerData
		err = json.Unmarshal(respBytes, &output)
		Expect(err).NotTo(HaveOccurred())

		hostIP, _, err := net.ParseCIDR(output[0].HostIP)
		Expect(err).NotTo(HaveOccurred())

		Expect(output[0].ID).To(Equal(containerID))
		Expect(output[0].IP).To(Equal("192.168.1.2"))
		Expect(output[0].MAC).To(MatchRegexp("[[:xdigit:]]{2}:[[:xdigit:]]{2}:[[:xdigit:]]{2}:[[:xdigit:]]{2}:[[:xdigit:]]{2}:[[:xdigit:]]{2}"))
		Expect(hostIP).NotTo(BeNil())

		By("invoking the vxlan CNI plugin with the DELETE action")
		delCmd, err := buildCNICmd("DEL", netConfig, containerNS, containerID, sandboxRepoDir, serverURL)
		Expect(err).NotTo(HaveOccurred())
		pluginSession, err = gexec.Start(delCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(pluginSession, DEFAULT_TIMEOUT).Should(gexec.Exit(0))

		By("checking that the daemon now has no containers saved")
		resp, err = http.Get(url)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()

		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		respBytes, err = ioutil.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())

		Expect(respBytes).To(MatchJSON(`[]`))
	})
})
