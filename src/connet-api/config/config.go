package config

import (
	"encoding/json"
	"fmt"
	"io"
	"lib/db"
	"os"
)

type Config struct {
	ListenHost string    `json:"listen_host"`
	ListenPort int       `json:"listen_port"`
	Database   db.Config `json: "database"`
}

func (c Config) Marshal(output io.Writer) error {
	encoder := json.NewEncoder(output)

	err := encoder.Encode(&c)
	if err != nil {
		return fmt.Errorf("json encode: %s", err) // not tested
	}

	return nil
}

func Unmarshal(input io.Reader) (*Config, error) {
	c := &Config{}
	decoder := json.NewDecoder(input)

	err := decoder.Decode(c)
	if err != nil {
		return nil, fmt.Errorf("json decode: %s", err)
	}

	return c, nil
}

func ParseConfigFile(configFilePath string) (*Config, error) {
	if configFilePath == "" {
		return nil, fmt.Errorf("missing config file path")
	}

	configFile, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	daemonConfig, err := Unmarshal(configFile)
	if err != nil {
		return nil, fmt.Errorf("parsing config: %s", err)
	}

	return daemonConfig, nil
}
