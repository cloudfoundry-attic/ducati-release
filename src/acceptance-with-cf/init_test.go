package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var (
	appDir string
	config helpers.Config
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)

	BeforeSuite(func() {
		config = helpers.LoadConfig()

		Expect(cf.Cf("api", "--skip-ssl-validation", config.ApiEndpoint).Wait(Timeout_Push)).To(gexec.Exit(0))

		cmd := exec.Command("cf", "auth", config.AdminUser, config.AdminPassword)
		sess, err := gexec.Start(cmd, nil, nil)
		Expect(err).NotTo(HaveOccurred())
		Eventually(sess.Wait(Timeout_Push)).Should(gexec.Exit(0))

		appDir = os.Getenv("APP_DIR")
		Expect(appDir).NotTo(BeEmpty())

		// create binary
		os.Setenv("GOOS", "linux")
		os.Setenv("GOARCH", "amd64")
		binaryPath, err := gexec.Build(appDir)
		Expect(err).NotTo(HaveOccurred())
		err = os.Rename(binaryPath, filepath.Join(appDir, "proxy"))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterSuite(func() {
		// remove binary
		err := os.Remove(filepath.Join(appDir, "proxy"))
		Expect(err).NotTo(HaveOccurred())
	})

	RunSpecs(t, "Acceptance Suite")
}

func pushApp(appName string) {
	Expect(cf.Cf(
		"push", appName,
		"-p", appDir,
		"-f", filepath.Join(appDir, "manifest.yml"),
		"-c", "./proxy",
		"-b", "binary_buildpack",
	).Wait(Timeout_Push)).To(gexec.Exit(0))
}
