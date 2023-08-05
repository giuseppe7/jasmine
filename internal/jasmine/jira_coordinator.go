package jasmine

import (
	"log"
	"strconv"

	"github.com/andygrunwald/go-jira"
	"github.com/prometheus/client_golang/prometheus"
)

type JiraCoordinator struct {
	config           JasmineConfig
	client           *jira.Client
	jqlExecHistogram *prometheus.HistogramVec
}

func NewJiraCoordinator(config JasmineConfig) (*JiraCoordinator, error) {
	// Create the jira client.
	tp := jira.BasicAuthTransport{
		Username: config.JiraUser,
		Password: config.JiraApiKey,
	}
	client, err := jira.NewClient(tp.Client(), config.JiraServer)
	if err != nil {
		return nil, err
	}

	// No issues creating the client, create the coordinator.
	jiraCoordinator := new(JiraCoordinator)
	jiraCoordinator.client = client
	jiraCoordinator.config = config
	jiraCoordinator.jqlExecHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: ApplicationNamespace,
			Name:      "jql_query_exec_seconds",
			Help:      "Histogram for jql execution results in seconds.",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 20, 30, 60},
		},
		[]string{"name", "exitCode"},
	)
	prometheus.MustRegister(jiraCoordinator.jqlExecHistogram)

	return jiraCoordinator, err
}

func (jiraCoordinator *JiraCoordinator) DoWork() {
	workStatusChannel := make(chan JiraWorkStatus, len(jiraCoordinator.config.Queries))

	for _, query := range jiraCoordinator.config.Queries {
		worker, err := NewJiraWorker(jiraCoordinator.client, query.Name, query.JQL, query.Attributes)
		if err != nil {
			log.Printf("Error in creating '%s' jira worker: %v\n", query.Name, err)
		} else {
			go worker.DoWork(workStatusChannel)
		}
	}

	for {
		workStatus := <-workStatusChannel
		jiraCoordinator.jqlExecHistogram.WithLabelValues(workStatus.name, strconv.FormatBool(workStatus.successful)).Observe(workStatus.elapsedTime)
		if !workStatus.successful {
			log.Printf("jira worker '%s' had status success of %t. %v\n", workStatus.name, workStatus.successful, workStatus.err)
		}
	}

}
