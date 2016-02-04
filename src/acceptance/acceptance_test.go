package acceptance_test

import (
	"bytes"

	"github.com/cloudfoundry-incubator/garden"
	"github.com/cloudfoundry-incubator/garden/client"
	"github.com/cloudfoundry-incubator/garden/client/connection"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Guardian integration with Ducati", func() {
	Describe("container creation", func() {
		It("should create interfaces", func() {
			gardenAddress := "10.244.16.2:7777"
			gardenClient := client.New(connection.New("tcp", gardenAddress))

			container, err := gardenClient.Create(garden.ContainerSpec{
				Network: "test-network",
			})
			Expect(err).NotTo(HaveOccurred())

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

			err = gardenClient.Destroy(container.Handle())
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
