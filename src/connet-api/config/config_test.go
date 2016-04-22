package config_test

import (
	"connet-api/config"
	"io/ioutil"
	"lib/db"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("config", func() {
	var fixtureConfig config.Config

	BeforeEach(func() {
		fixtureConfig = config.Config{
			ListenHost: "127.0.0.1",
			ListenPort: 1234,
			Database: db.Config{
				Host:     "example.com",
				Port:     9953,
				Username: "bob",
				Password: "secret",
				Name:     "database1",
				SSLMode:  "false",
			},
		}
	})

	Describe("loading config from a file", func() {
		It("returns the parsed config", func() {
			configFile, err := ioutil.TempFile("", "config")
			Expect(err).NotTo(HaveOccurred())

			Expect(fixtureConfig.Marshal(configFile)).To(Succeed())
			configFile.Close()

			conf, err := config.ParseConfigFile(configFile.Name())
			Expect(err).NotTo(HaveOccurred())

			Expect(conf).To(Equal(&fixtureConfig))
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
