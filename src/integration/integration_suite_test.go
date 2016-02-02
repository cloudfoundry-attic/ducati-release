package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/appc/cni/pkg/types"
	"github.com/cloudfoundry-incubator/ducati-cni-plugins/lib/namespace"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var pathToVxlan, pathToDaemon, cniPath string

type Config struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Network     string `json:"network"`
	HostNetwork string `json:"host_network"`
	IPAM        IPAM   `json:"ipam,omitempty"`
}

type IPAM struct {
	Type   string              `json:"type,omitempty"`
	Subnet string              `json:"subnet,omitempty"`
	Routes []map[string]string `json:"routes,omitempty"`
}

type paths struct {
	VXLAN    string `json:"vxlan"`
	FAKEIPAM string `json:"fake_ipam"`
	DAEMON   string `json:"daemon"`
}

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		if runtime.GOOS != "linux" {
			Skip("Cannot run suite for non linux platform: " + runtime.GOOS)
		}

		// race detector doesn't work with cgo in go 1.5
		vxlan, err := gexec.Build("github.com/cloudfoundry-incubator/ducati-cni-plugins/cmd/vxlan")
		Expect(err).NotTo(HaveOccurred())

		fakeIpam, err := gexec.Build("github.com/cloudfoundry-incubator/ducati-cni-plugins/fake_plugins")
		Expect(err).NotTo(HaveOccurred())

		pathToDaemon, err := gexec.Build("github.com/cloudfoundry-incubator/ducati-daemon/cmd/ducatid/")
		Expect(err).NotTo(HaveOccurred())

		result, err := json.Marshal(paths{
			VXLAN:    vxlan,
			FAKEIPAM: fakeIpam,
			DAEMON:   pathToDaemon,
		})
		Expect(err).NotTo(HaveOccurred())

		return result
	},
	func(result []byte) {
		var paths paths
		err := json.Unmarshal(result, &paths)
		Expect(err).NotTo(HaveOccurred())

		vxlanBinDir := filepath.Dir(paths.VXLAN)
		fakeIpamDir := filepath.Dir(paths.FAKEIPAM)
		pathToDaemon = paths.DAEMON

		cniPath = fmt.Sprintf("%s%c%s", vxlanBinDir, os.PathListSeparator, fakeIpamDir)
		pathToVxlan = paths.VXLAN
	},
)

var _ = SynchronizedAfterSuite(func() {
	return
}, func() {
	gexec.CleanupBuildArtifacts()
})

func buildCNICmd(operation string, netConfig Config, containerNS namespace.Namespace,
	containerID, sandboxRepoDir string) (*exec.Cmd, error) {

	input, err := json.Marshal(netConfig)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(pathToVxlan)
	cmd.Stdin = bytes.NewReader(input)
	fakeIPAMResponse := &types.Result{
		IP4: &types.IPConfig{
			IP: net.IPNet{
				IP:   net.ParseIP("192.168.1.3"),
				Mask: net.ParseIP("192.168.1.1").DefaultMask(),
			},
			Gateway: net.ParseIP("192.168.1.1"),
			Routes: []types.Route{
				{
					Dst: net.IPNet{
						IP:   net.ParseIP("0.0.0.0"),
						Mask: net.IPv4Mask(0, 0, 0, 0),
					},
				},
			},
		},
	}
	fakeIPAMResponseBytes, err := json.Marshal(fakeIPAMResponse)
	if err != nil {
		return nil, err
	}
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("CNI_COMMAND=%s", operation),
		fmt.Sprintf("CNI_CONTAINERID=%s", containerID),
		fmt.Sprintf("CNI_PATH=%s", cniPath),
		fmt.Sprintf("CNI_NETNS=%s", containerNS.Path()),
		fmt.Sprintf("CNI_IFNAME=%s", "vx-eth0"),
		fmt.Sprintf("DUCATI_OS_SANDBOX_REPO=%s", sandboxRepoDir),
		fmt.Sprintf("FAKE_IPAM_RESPONSE=%s", string(fakeIPAMResponseBytes)),
	)

	return cmd, nil
}
