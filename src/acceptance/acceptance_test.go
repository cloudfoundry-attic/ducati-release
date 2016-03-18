package acceptance_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/garden/client/connection"

	ducati_client "github.com/cloudfoundry-incubator/ducati-daemon/client"
	"github.com/cloudfoundry-incubator/ducati-daemon/models"
	garden_client "github.com/cloudfoundry-incubator/garden/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Guardian integration with Ducati", func() {
	const networkName = "vni-1"

	Context("when there is one garden server", func() {
		var gardenClient1 garden.Client
		var ducatiClient1 *ducati_client.DaemonClient
		var gardenContainer garden.Container
		var ducatiContainer *models.Container

		BeforeEach(func() {
			gardenAddress := fmt.Sprintf("%s:7777", gardenServer1)
			gardenClient1 = garden_client.New(connection.New("tcp", gardenAddress))
			ducatiClient1 = ducati_client.New(fmt.Sprintf("http://%s:4001", gardenServer1), http.DefaultClient)

			var err error
			gardenContainer, err = gardenClient1.Create(garden.ContainerSpec{
				Network: networkName,
			})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() error {
				containers, err := ducatiClient1.ListNetworkContainers(networkName)
				if err != nil {
					return err
				}

				for _, c := range containers {
					if c.ID == gardenContainer.Handle() {
						ducatiContainer = &c
						return nil
					}
				}

				return errors.New("container not found")
			}, "5s").Should(Succeed())
		})

		AfterEach(func() {
			err := gardenClient1.Destroy(gardenContainer.Handle())
			Expect(err).NotTo(HaveOccurred())

			ducatiClient1 := ducati_client.New(fmt.Sprintf("http://%s:4001", gardenServer1), http.DefaultClient)
			Eventually(func() ([]models.Container, error) {
				containers, err := ducatiClient1.ListNetworkContainers(networkName)
				return containers, err
			}, "5s").Should(BeEmpty())
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

			process, err := gardenContainer.Run(ifconfigProcess, ifconfigProcessIO)
			Expect(err).NotTo(HaveOccurred())
			Eventually(process.Wait).Should(Equal(0))

			output := stdout.String()
			Expect(output).To(ContainSubstring("eth0"))
			Expect(output).To(ContainSubstring("eth1"))
		})

		Context("when containers share a network", func() {
			var gardenContainer2 garden.Container
			var ducatiContainer2 *models.Container

			BeforeEach(func() {
				var err error
				gardenContainer2, err = gardenClient1.Create(garden.ContainerSpec{
					Network: networkName,
				})
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() error {
					containers, err := ducatiClient1.ListNetworkContainers(networkName)
					if err != nil {
						return err
					}

					for _, c := range containers {
						if c.ID == gardenContainer2.Handle() {
							ducatiContainer2 = &c
							return nil
						}
					}

					return errors.New("container not found")
				}, "5s").Should(Succeed())
			})

			AfterEach(func() {
				err := gardenClient1.Destroy(gardenContainer2.Handle())
				Expect(err).NotTo(HaveOccurred())
			})

			It("connects the containers", func() {
				By("pinging from container 1 to container 2")
				pingContainer2 := garden.ProcessSpec{
					Path: "/bin/ping",
					Args: []string{"-c3", ducatiContainer2.IP},
					User: "root",
				}

				process, err := gardenContainer.Run(pingContainer2, garden.ProcessIO{})
				Expect(err).NotTo(HaveOccurred())
				Eventually(process.Wait).Should(Equal(0))

				By("pinging from container 2 to container 1")
				pingContainer1 := garden.ProcessSpec{
					Path: "/bin/ping",
					Args: []string{"-c3", ducatiContainer.IP},
					User: "root",
				}

				process, err = gardenContainer2.Run(pingContainer1, garden.ProcessIO{})
				Expect(err).NotTo(HaveOccurred())
				Eventually(process.Wait).Should(Equal(0))
			})
		})
	})

	Context("when there are two garden servers", func() {
		var (
			gardenClient1 garden.Client
			gardenClient2 garden.Client

			ducatiClient1 *ducati_client.DaemonClient
			ducatiClient2 *ducati_client.DaemonClient

			gardenContainer  garden.Container
			gardenContainer2 garden.Container

			ducatiContainer  *models.Container
			ducatiContainer2 *models.Container
		)

		BeforeEach(func() {
			gardenAddress1 := fmt.Sprintf("%s:7777", gardenServer1)
			gardenAddress2 := fmt.Sprintf("%s:7777", gardenServer2)

			gardenClient1 = garden_client.New(connection.New("tcp", gardenAddress1))
			gardenClient2 = garden_client.New(connection.New("tcp", gardenAddress2))

			ducatiClient1 = ducati_client.New(fmt.Sprintf("http://%s:4001", gardenServer1), http.DefaultClient)
			ducatiClient2 = ducati_client.New(fmt.Sprintf("http://%s:4001", gardenServer2), http.DefaultClient)

			var err error
			gardenContainer, err = gardenClient1.Create(garden.ContainerSpec{
				Network: networkName,
			})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() error {
				containers, err := ducatiClient1.ListNetworkContainers(networkName)
				if err != nil {
					return err
				}

				for _, c := range containers {
					if c.ID == gardenContainer.Handle() {
						ducatiContainer = &c
						return nil
					}
				}

				return errors.New("container not found")
			}, "5s").Should(Succeed())

			gardenContainer2, err = gardenClient2.Create(garden.ContainerSpec{
				Network: networkName,
			})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() error {
				containers, err := ducatiClient1.ListNetworkContainers(networkName)
				if err != nil {
					return err
				}

				for _, c := range containers {
					if c.ID == gardenContainer2.Handle() {
						ducatiContainer2 = &c
						return nil
					}
				}

				return errors.New("container not found")
			}, "5s").Should(Succeed())
		})

		AfterEach(func() {
			err := gardenClient1.Destroy(gardenContainer.Handle())
			Expect(err).NotTo(HaveOccurred())

			err = gardenClient2.Destroy(gardenContainer2.Handle())
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() ([]models.Container, error) {
				containers, err := ducatiClient1.ListNetworkContainers(networkName)
				return containers, err
			}, "5s").Should(BeEmpty())
		})

		It("should share container metadata across the deployment", func() {
			containersList1, err := ducatiClient1.ListNetworkContainers(networkName)
			Expect(err).NotTo(HaveOccurred())

			containersList2, err := ducatiClient2.ListNetworkContainers(networkName)
			Expect(err).NotTo(HaveOccurred())

			Expect(containersList1).To(ConsistOf(containersList2))
		})

		It("connects the containers", func() {
			By("pinging from container 1 to container 2")
			pingContainer2 := garden.ProcessSpec{
				Path: "/bin/ping",
				Args: []string{"-c3", ducatiContainer2.IP},
				User: "root",
			}

			process, err := gardenContainer.Run(pingContainer2, garden.ProcessIO{})
			Expect(err).NotTo(HaveOccurred())
			Eventually(process.Wait).Should(Equal(0))

			By("pinging from container 2 to container 1")
			pingContainer1 := garden.ProcessSpec{
				Path: "/bin/ping",
				Args: []string{"-c3", ducatiContainer.IP},
				User: "root",
			}

			process, err = gardenContainer2.Run(pingContainer1, garden.ProcessIO{})
			Expect(err).NotTo(HaveOccurred())
			Eventually(process.Wait).Should(Equal(0))
		})
	})
})
