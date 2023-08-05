package jasmine

import (
	"errors"

	"github.com/spf13/viper"
)

type JasmineConfig struct {
	JiraUser   string `yaml:"jiraUser"`
	JiraApiKey string `yaml:"jiraApiKey"`
	JiraServer string `yaml:"jiraServer"`
	Queries    []struct {
		Name       string   `yaml:"name"`
		JQL        string   `yaml:"jql"`
		Attributes []string `yaml:"attributes"`
	}
}

func ReadInConfig(configFile string) (JasmineConfig, error) {
	var config JasmineConfig
	viper.SetEnvPrefix("JASMINE")

	err := viper.BindEnv("jiraServer")
	if err != nil {
		return config, err
	}

	err = viper.BindEnv("jiraUser")
	if err != nil {
		return config, err
	}

	err = viper.BindEnv("jiraApiKey")
	if err != nil {
		return config, err
	}

	viper.AutomaticEnv()

	viper.AddConfigPath(".")
	viper.SetConfigFile(configFile)
	err = viper.ReadInConfig()
	if err != nil {
		return config, err
	}
	err = viper.Unmarshal(&config)

	// Jasmine configuration minimally needs to know Jira destination and credentials.
	if len(config.JiraServer) == 0 {
		return config, errors.New("config did not determine jira server")
	} else if len(config.JiraUser) == 0 {
		return config, errors.New("config did not determine jira user")
	} else if len(config.JiraApiKey) == 0 {
		return config, errors.New("config did not determine jira api key")
	}

	return config, err
}
