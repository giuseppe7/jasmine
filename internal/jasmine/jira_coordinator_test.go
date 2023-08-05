package jasmine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateNewJiraCoordinator(t *testing.T) {
	var config JasmineConfig
	var exampleConfigFile = "../../configs/jasmine/config.yaml"
	config, err := ReadInConfig(exampleConfigFile)
	require.NoError(t, err)

	_, err = NewJiraCoordinator(config)
	require.NoError(t, err)
}
