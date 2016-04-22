package config_test

import (
	"connet-api/config"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const fixtureJSON = `
{
	"listen_host": "0.0.0.0",
	"listen_port": 4001
}
`

var _ = Describe("config", func() {
	var fixtureConfig config.Config

	BeforeEach(func() {
		Expect(true).To(BeTrue())
		fixtureConfig = config.Config{
			ListenHost: "127.0.0.1",
			ListenPort: 1234,
		}
	})

	Describe("loading config from a file", func() {
		It("returns the parsed and validated config", func() {
			configSource := config.Config{
				ListenHost: "127.0.0.1",
				ListenPort: 4001,
			}

			configFile, err := ioutil.TempFile("", "config")
			Expect(err).NotTo(HaveOccurred())

			Expect(configSource.Marshal(configFile)).To(Succeed())
			configFile.Close()

			conf, err := config.ParseConfigFile(configFile.Name())
			Expect(err).NotTo(HaveOccurred())

			Expect(conf).To(Equal(&config.Config{
				ListenHost: "127.0.0.1",
				ListenPort: 4001,
			}))
		})

		Context("when configFilePath is not present", func() {
			It("returns an error", func() {
				_, err := config.ParseConfigFile("")
				Expect(err).To(MatchError("missing config file path"))
			})
		})

		Context("when the config file cannot be opened", func() {
			It("returns an error", func() {
				_, err := config.ParseConfigFile("some-path")
				Expect(err).To(MatchError("open some-path: no such file or directory"))
			})
		})

		Context("when config file contents cannot be unmarshaled", func() {
			It("returns an error", func() {
				_, err := config.ParseConfigFile("/dev/null")
				Expect(err).To(MatchError("parsing config: json decode: EOF"))
			})
		})
	})
})
