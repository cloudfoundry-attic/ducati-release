package acceptance_test

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/garden/client/connection"

	ducati_client "github.com/cloudfoundry-incubator/ducati-daemon/client"
	"github.com/cloudfoundry-incubator/ducati-daemon/models"
	garden_client "github.com/cloudfoundry-incubator/garden/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Guardian integration with Ducati", func() {
	Describe("container creation", func() {
		var gardenClient1 garden_client.Client
		var container garden.Container

		BeforeEach(func() {
			gardenAddress := fmt.Sprintf("%s:7777", gardenServer1)

			gardenClient1 = garden_client.New(connection.New("tcp", gardenAddress))

			var err error
			container, err = gardenClient1.Create(garden.ContainerSpec{
				Network: "test-network",
			})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := gardenClient1.Destroy(container.Handle())
			Expect(err).NotTo(HaveOccurred())

			daemonClient1 := ducati_client.New(fmt.Sprintf("http://%s:4001", gardenServer1), http.DefaultClient)
			Eventually(func() ([]models.Container, error) {
				containers, err := daemonClient1.ListContainers()
				return containers, err
			}, "5s").Should(BeEmpty())

			time.Sleep(3 * time.Second)
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

		Context("when there are two garden servers", func() {
			var gardenClient2 garden.Client
			var container2 garden.Container

			var ducatiClient1, ducatiClient2 *ducati_client.DaemonClient

			BeforeEach(func() {
				gardenAddress2 := fmt.Sprintf("%s:7777", gardenServer2)

				gardenClient2 = garden_client.New(connection.New("tcp", gardenAddress2))

				ducatiClient1 = ducati_client.New(fmt.Sprintf("http://%s:4001", gardenServer1), http.DefaultClient)
				ducatiClient2 = ducati_client.New(fmt.Sprintf("http://%s:4001", gardenServer2), http.DefaultClient)

				var err error
				container2, err = gardenClient2.Create(garden.ContainerSpec{
					Network: "test-network",
				})
				Expect(err).NotTo(HaveOccurred())
			})
			AfterEach(func() {
				err := gardenClient2.Destroy(container2.Handle())
				Expect(err).NotTo(HaveOccurred())
			})

			It("should share container metadata across the deployment", func() {
				containersList1, err := ducatiClient1.ListContainers()
				Expect(err).NotTo(HaveOccurred())

				containersList2, err := ducatiClient2.ListContainers()
				Expect(err).NotTo(HaveOccurred())

				Expect(containersList1).To(ConsistOf(containersList2))
			})
		})
	})
})
