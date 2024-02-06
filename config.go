package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/charmbracelet/log"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Addr           string   `yaml:"addr"`
	TrustedKeys    []string `yaml:"trusted_keys"`
	LogLevel       string   `yaml:"log_level"`
	authorizedKeys string   `yaml:"-"` //Private field for user process keys
}

func NewConfig() (*Config, error) {
	path := flag.String("config", "", "path to the config file")
	debug := flag.Bool("debug", false, "debug log level")
	flag.Parse()
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	config := &Config{}
	if err := loadConfigFromPath(config, *path); err != nil {
		return nil, err
	}
	return config, setDefaults(config) //return the config and error from the default settings
}

// Load the configuration from the file
func loadConfigFromPath(config *Config, path string) error {
	if path != "" {
		yamlFile, err := os.Open(path)
		if err == nil {
			log.Info("loading configuration from file", "path", path)
			defer yamlFile.Close()
			dec := yaml.NewDecoder(yamlFile)
			dec.SetStrict(true)
			if err := dec.Decode(config); err != nil {
				return fmt.Errorf("can't parse configuration file %s : %s", path, err)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("can't open configuration file %s : %s", path, err)
		}
	}
	return nil
}

/*
Set default values for the configuration
*/
func setDefaults(config *Config) error {
	//Set default addr to all listeners port 1337
	if config.Addr == "" {
		config.Addr = "0.0.0.0:1337"
	}
	loadCurrentUserAuthorizedKeys(config)
	return nil
}

func loadCurrentUserAuthorizedKeys(config *Config) {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Debug("user home directory is not configured", "error", err)
		return
	}
	authorizedKeysPath := path.Join(home, ".ssh", "authorized_keys")
	if _, err := os.Stat(authorizedKeysPath); err != nil {
		log.Debug("user has no authorized keys file", "error", err)
		return
	}
	config.authorizedKeys = authorizedKeysPath
}
