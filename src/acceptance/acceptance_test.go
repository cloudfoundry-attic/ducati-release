package acceptance_test

import (
	"bytes"
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/garden/client"
	"github.com/cloudfoundry-incubator/garden/client/connection"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Guardian integration with Ducati", func() {
	Describe("container creation", func() {
		var gardenClient client.Client
		var container garden.Container

		BeforeEach(func() {
			gardenServer := os.Getenv("GARDEN_SERVER_1")
			if gardenServer == "" {
				gardenServer = "10.244.16.2"
			}
			gardenAddress := fmt.Sprintf("%s:7777", gardenServer)

			gardenClient = client.New(connection.New("tcp", gardenAddress))

			var err error
			container, err = gardenClient.Create(garden.ContainerSpec{
				Network: "test-network",
			})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := gardenClient.Destroy(container.Handle())
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create interfaces", func() {
			ifconfigProcess := garden.ProcessSpec{
				Path: "/sbin/ifconfig",
				Args: []string{"-a"},
				User: "root",
			}

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			ifconfigProcessIO := garden.ProcessIO{
				Stdin:  &bytes.Buffer{},
				Stdout: stdout,
				Stderr: stderr,
			}

			process, err := container.Run(ifconfigProcess, ifconfigProcessIO)
			Expect(err).NotTo(HaveOccurred())
			Eventually(process.Wait).Should(Equal(0))

			output := stdout.String()
			Expect(output).To(ContainSubstring("eth0"))
			Expect(output).To(ContainSubstring("eth1"))
		})

		It("should define routes to for the overlay", func() {
			ifconfigProcess := garden.ProcessSpec{
				Path: "/sbin/ip",
				Args: []string{"route"},
				User: "root",
			}

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			ifconfigProcessIO := garden.ProcessIO{
				Stdin:  &bytes.Buffer{},
				Stdout: stdout,
				Stderr: stderr,
			}

			process, err := container.Run(ifconfigProcess, ifconfigProcessIO)
			Expect(err).NotTo(HaveOccurred())
			Eventually(process.Wait).Should(Equal(0))

			output := stdout.String()

			Expect(output).To(ContainSubstring("192.168.0.0/16 via 192.168."))
		})
	})
})
