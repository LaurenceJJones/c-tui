package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Addr string
}

func NewConfig(path string) (*Config, error) {
	//read path and parse yaml
	config := &Config{}
	if err := loadConfigFromPath(config, path); err != nil {
		return nil, err
	}
	setDefaults(config)
	return config, nil
}

func loadConfigFromPath(config *Config, path string) error {
	if path != "" {
		log.Info("loading configuration from file", "path", path)
		yamlFile, err := os.Open(path)
		if err == nil {
			defer yamlFile.Close()
			//process the yaml
			dec := yaml.NewDecoder(yamlFile)
			dec.SetStrict(true)
			if err := dec.Decode(config); err != nil {
				return fmt.Errorf("can't parse configuration file %s : %s", path, err)
			}
		}
	}
	return nil
}

/*
Set default values for the configuration
*/
func setDefaults(config *Config) {
	//Set default addr to all listeners port 1337
	if config.Addr == "" {
		config.Addr = "0.0.0.0:1337"
	}
}
