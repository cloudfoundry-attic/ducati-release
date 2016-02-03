package integration_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"

	"github.com/cloudfoundry-incubator/ducati-cni-plugins/lib/namespace"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("how the VXLAN plugin talks to the ducati daemon", func() {
	var (
		session     *gexec.Session
		address     string
		repoDir     string
		containerNS namespace.Namespace
		containerID string
		netConfig   Config
		serverURL   string
	)

	BeforeEach(func() {
		By("booting the daemon")
		address = fmt.Sprintf("127.0.0.1:%d", 4001+GinkgoParallelNode())
		serverURL = "http://" + address
		daemonCmd := exec.Command(pathToDaemon, "-listenAddr", address)
		var err error
		session, err = gexec.Start(daemonCmd, GinkgoWriter, GinkgoWriter)
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
			Type:    "vxlan",
			Network: "192.168.1.0/24",
			IPAM: IPAM{
				Type: "fake_plugins",
			},
		}
	})

	AfterEach(func() {
		Expect(containerNS.Destroy()).To(Succeed())
		Expect(os.RemoveAll(repoDir)).To(Succeed())

		Eventually(session.Terminate()).Should(gexec.Exit(0))
	})

	var serverIsAvailable = func() error {
		_, err := net.Dial("tcp", address)
		return err
	}

	It("should maintain the container state in the daemon", func() {
		url := fmt.Sprintf("http://%s/containers", address)
		Eventually(serverIsAvailable).Should(Succeed())

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
		session, err = gexec.Start(addCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		By("checking that the daemon now how the ID and IP of the container")
		resp, err = http.Get(url)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()

		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		respBytes, err = ioutil.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())

		Expect(respBytes).To(MatchJSON(fmt.Sprintf(`[{"id": "%s"}]`, containerID)))

		By("invoking the vxlan CNI plugin with the DELETE action")
		delCmd, err := buildCNICmd("DEL", netConfig, containerNS, containerID, repoDir, serverURL)
		Expect(err).NotTo(HaveOccurred())
		session, err = gexec.Start(delCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		By("checking that the daemon now has non containers saved")
		resp, err = http.Get(url)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()

		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		respBytes, err = ioutil.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())

		Expect(respBytes).To(MatchJSON(`[]`))
	})
})
