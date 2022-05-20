package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

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
	flag.StringVar(&flagConfigurationType, "config-type", "yaml", "The type of configuration file <json|yaml>")
	flag.StringVar(&flagConfigurationFileVersion, "config-version", "0.2", "The version of the configuration file")
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
	case flagConfigurationType == "json":
		switch {
		case flagConfigurationFileVersion == "0.1":
			var cfs ConfigurationFileSite

			err = json.Unmarshal(configurationFileBody, &cfs)

			cf.Sites = []ConfigurationFileSite{cfs}
		case flagConfigurationFileVersion == "0.2":
			err = json.Unmarshal(configurationFileBody, &cf)
		default:
			return nil, fmt.Errorf("%s %s", "Unsupported configuration file version:", flagConfigurationFileVersion)
		}

		if err != nil {
			return nil, fmt.Errorf("%s %s", "Error decoding configuration JSON:", err)
		}
	case flagConfigurationType == "yaml":
		err = yaml.Unmarshal(configurationFileBody, &cf)

		if err != nil {
			return nil, fmt.Errorf("%s %s", "Error decoding configuration YAML:", err)
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
