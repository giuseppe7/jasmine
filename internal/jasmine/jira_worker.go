package jasmine

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/prometheus/client_golang/prometheus"
)

const NilAttributeValue = "nil"

type JiraWorker struct {
	client            *jira.Client
	name              string
	jql               string
	attributes        []string
	queryResultCounts *prometheus.GaugeVec
}

func NewJiraWorker(client *jira.Client, name string, jql string, attributes []string) (*JiraWorker, error) {
	jiraWorker := new(JiraWorker)
	jiraWorker.client = client
	jiraWorker.name = name
	jiraWorker.jql = jql
	jiraWorker.attributes = attributes

	jiraWorker.queryResultCounts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: ApplicationNamespace,
			Name:      fmt.Sprintf("%s_query_results_gauge", jiraWorker.name),
			Help:      "Gauge for query results and attributes.",
		},
		[]string{"name", "attributeName", "attributeValue"},
	)
	prometheus.MustRegister(jiraWorker.queryResultCounts)

	return jiraWorker, nil
}

func (jiraWorker *JiraWorker) GetJiraUserDisplayName(user *jira.User) string {
	if user == nil {
		return NilAttributeValue
	} else {
		return user.DisplayName
	}
}

func (jiraWorker *JiraWorker) GetJiraIssueAttributeValue(issue jira.Issue, attribute string) (string, error) {
	// Loop through known items, then move to unknowns.
	if strings.ToLower(attribute) == "reporter" {
		return jiraWorker.GetJiraUserDisplayName(issue.Fields.Reporter), nil
	} else if strings.ToLower(attribute) == "assignee" {
		return jiraWorker.GetJiraUserDisplayName(issue.Fields.Assignee), nil
	} else if strings.ToLower(attribute) == "issuetype" {
		return issue.Fields.Type.Description, nil
	}

	// Now it gets messy.
	if strings.HasPrefix(strings.ToLower(attribute), "customfield") {
		// Check if the attribute exist in the Unknowns field.
		if customField, ok := issue.Fields.Unknowns[attribute]; ok {
			// Check if this is null for the found attribute.
			if customField == nil {
				return NilAttributeValue, nil
			}
			// Now see if we can unmarshal it.
			customFieldMap, ok := customField.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("unmarshaling custom field: %v", customFieldMap)
			} else {
				value := fmt.Sprintf("%v", customFieldMap["value"])
				return value, nil
			}
		}
	}

	return "", errors.New("unexpected scenario in getting jira issue attribute value")
}

func (jiraWorker *JiraWorker) DoWork(workStatusChannel chan<- JiraWorkStatus) {
	const sleepInterval = 60

	for {
		// JiraWorkStatus to encapsulate the iteration's result.
		workStatus := JiraWorkStatus{
			name:        jiraWorker.name,
			count:       0,
			elapsedTime: 0,
			successful:  false,
			err:         nil,
		}

		// Get the Jira issues.
		startTime := time.Now()
		issues, err := jiraWorker.GetJiraIssues()
		workStatus.elapsedTime = time.Since(startTime).Seconds()
		if err != nil {
			workStatus.successful = false
			workStatus.err = err
		} else {
			workStatus.count = len(issues)
			workStatus.successful = true
		}
		workStatusChannel <- workStatus

		// Reset the attribute name map to no values.
		attrNamesMap := make(map[string]map[string]int)
		for _, attrName := range jiraWorker.attributes {
			nestedMap := make(map[string]int)
			attrNamesMap[attrName] = nestedMap
		}

		// Iterate through the issues to collect stats on attributes.
		for _, issue := range issues {
			for attrName := range attrNamesMap {
				attrValue, err := jiraWorker.GetJiraIssueAttributeValue(issue, attrName)
				if err != nil {
					log.Printf("jira worker '%s' for %s errored with %v.\n", jiraWorker.name, attrName, err)
				} else {
					nestedMap := attrNamesMap[attrName]
					if value, exists := nestedMap[attrValue]; exists {
						nestedMap[attrValue] = value + 1
					} else {
						nestedMap[attrValue] = 1
					}
				}
			}
		}

		// Summarize it via metrics endpoint.
		for attrName := range attrNamesMap {
			attrNameValueMap := attrNamesMap[attrName]
			for attrValue := range attrNameValueMap {
				log.Printf("jira worker '%s' has %s with %d.\n", jiraWorker.name, attrValue, attrNameValueMap[attrValue])
				jiraWorker.queryResultCounts.WithLabelValues(jiraWorker.name, attrName, attrValue).Set(float64(attrNameValueMap[attrValue]))
			}
		}
		time.Sleep(sleepInterval * time.Second)
	}
}

func (jiraWorker *JiraWorker) GetJiraIssues() ([]jira.Issue, error) {
	var issues []jira.Issue

	appendFunc := func(i jira.Issue) (err error) {
		issues = append(issues, i)
		return err
	}

	err := jiraWorker.client.Issue.SearchPages(jiraWorker.jql, nil, appendFunc)
	return issues, err
}
