package jasmine

import (
	"testing"

	"github.com/andygrunwald/go-jira"
	"github.com/stretchr/testify/require"
)

func TestGetResults(t *testing.T) {
	var config JasmineConfig
	var exampleConfigFile = "../../configs/jasmine/config.yaml"
	config, err := ReadInConfig(exampleConfigFile)
	require.NoError(t, err)

	tp := jira.BasicAuthTransport{
		Username: config.JiraUser,
		Password: config.JiraApiKey,
	}
	client, err := jira.NewClient(tp.Client(), config.JiraServer)
	require.NoError(t, err)

	jiraWorker, err := NewJiraWorker(client, config.Queries[0].Name, config.Queries[0].JQL, config.Queries[0].Attributes)
	require.NoError(t, err)
	require.NotEmpty(t, jiraWorker)
	require.Equal(t, config.Queries[0].JQL, jiraWorker.jql)
	require.Equal(t, config.Queries[0].Attributes, jiraWorker.attributes)

	issues, err := jiraWorker.GetJiraIssues()
	require.NoError(t, err)
	require.NotNil(t, issues)

	// Reset the attribute name map to no values.
	attrNamesMap := make(map[string]map[string]int)
	for _, attrName := range jiraWorker.attributes {
		nestedMap := make(map[string]int)
		attrNamesMap[attrName] = nestedMap
	}

	// Iterate through the issues to collect stats on attributes.
	for _, issue := range issues {
		for attrName := range attrNamesMap {
			_, err := jiraWorker.GetJiraIssueAttributeValue(issue, attrName)
			require.NoError(t, err, attrName)
		}
	}

}
