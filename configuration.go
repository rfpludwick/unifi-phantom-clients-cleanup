package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigurationFileVersionDetect struct {
	Version string
}

type ConfigurationFile struct {
	Version          string
	HttpLogDirectory string `yaml:"httpLogDirectory"`
	Sites            []ConfigurationFileSite
}

type ConfigurationFileSite struct {
	Host                string
	ApiHost             string
	Site                string
	Username            string
	Password            string
	Udmp                bool
	ValidateCertificate bool `yaml:"validateCertificate"`
}

var (
	flagShowHelp                 bool
	flagVerboseOutput            bool
	flagConfigurationFilename    string
	flagConfigurationType        string
	flagConfigurationFileVersion string
)

func initConfiguration() {
	flag.BoolVar(&flagShowHelp, "help", false, "Show help")
	flag.BoolVar(&flagVerboseOutput, "verbose", false, "Show verbose output")
	flag.StringVar(&flagConfigurationFilename, "config", "conf/configuration.yaml", "Path to the configuration file to load")
	flag.StringVar(&flagConfigurationType, "config-type", "yaml", "The type of configuration file <yaml|json>")
	flag.StringVar(&flagConfigurationFileVersion, "config-version", "detect", "The version of the configuration file")
}

func processConfiguration() (*ConfigurationFile, error) {
	flag.Parse()

	if flagShowHelp {
		flag.Usage()

		os.Exit(0)
	}

	configurationFileBody, err := os.ReadFile(flagConfigurationFilename)

	if err != nil {
		return nil, fmt.Errorf("%s %s", "Error reading configuration file:", err)
	}

	var cf ConfigurationFile

	switch {
	case flagConfigurationType == "yaml":
		if flagConfigurationFileVersion == "detect" {
			var cfvd ConfigurationFileVersionDetect

			err = yaml.Unmarshal(configurationFileBody, &cfvd)

			if err != nil {
				return nil, fmt.Errorf("%s %s", "Error decoding configuration YAML for version detection:", err)
			}

			flagConfigurationFileVersion = cfvd.Version
		}

		switch {
		case flagConfigurationFileVersion == "0.2":
			err = processConfigurationYaml02(configurationFileBody, &cf)
		case flagConfigurationFileVersion == "0.1":
			return nil, fmt.Errorf("%s %s", "YAML does not support configuration file verion 0.1:", flagConfigurationFileVersion)
		default:
			return nil, fmt.Errorf("%s %s", "Unsupported YAML configuration file version:", flagConfigurationFileVersion)
		}

		if err != nil {
			return nil, fmt.Errorf("%s %s", "Error decoding configuration YAML:", err)
		}
	case flagConfigurationType == "json":
		if flagConfigurationFileVersion == "detect" {
			var cfvd ConfigurationFileVersionDetect

			err = json.Unmarshal(configurationFileBody, &cfvd)

			if err != nil {
				return nil, fmt.Errorf("%s %s", "Error decoding configuration JSON for version detection:", err)
			}

			flagConfigurationFileVersion = cfvd.Version
		}

		switch {
		case flagConfigurationFileVersion == "0.2":
			err = processConfigurationJson02(configurationFileBody, &cf)
		case flagConfigurationFileVersion == "0.1":
			err = processConfigurationJson01(configurationFileBody, &cf)
		default:
			return nil, fmt.Errorf("%s %s", "Unsupported JSON configuration file version:", flagConfigurationFileVersion)
		}

		if err != nil {
			return nil, fmt.Errorf("%s %s", "Error decoding configuration JSON:", err)
		}
	}

	if cf.HttpLogDirectory != "" {
		_, err := os.Stat(cf.HttpLogDirectory)

		if err != nil {
			if os.IsNotExist(err) {
				err := os.Mkdir(cf.HttpLogDirectory, 0755)

				if err != nil {
					return nil, fmt.Errorf("%s %s", "Error creating HTTP log directory:", err)
				}
			} else {
				return nil, fmt.Errorf("%s %s", "Error STATing HTTP log directory:", err)
			}
		}
	}

	return &cf, nil
}

func processConfigurationYaml02(configurationFileBody []byte, cf *ConfigurationFile) error {
	return yaml.Unmarshal(configurationFileBody, &cf)
}

func processConfigurationJson02(configurationFileBody []byte, cf *ConfigurationFile) error {
	return json.Unmarshal(configurationFileBody, &cf)
}

func processConfigurationJson01(configurationFileBody []byte, cf *ConfigurationFile) error {
	var cfs ConfigurationFileSite

	err := json.Unmarshal(configurationFileBody, &cfs)

	if err != nil {
		return err
	}

	cf.Sites = []ConfigurationFileSite{cfs}

	return nil
}
