package jasmine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJasmineConfig(t *testing.T) {
	var config JasmineConfig
	var exampleConfigFile = "../../configs/jasmine/config.yaml"
	config, err := ReadInConfig(exampleConfigFile)

	require.NoError(t, err)
	require.NotEmpty(t, config.JiraServer, "expected Jira Server in config")
	require.NotEmpty(t, config.JiraUser, "expected Jira User in config")
	require.NotEmpty(t, config.JiraApiKey, "expectd Jira API Key in config")
	require.GreaterOrEqual(t, len(config.Queries), 1)
}
