package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/giuseppe7/jasmine/internal/jasmine"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var version string // Variable to be set by the Go linker at build time.

// Set up observability with Prometheus handler for metrics.
func initObservability() {

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		_ = http.ListenAndServe(":2112", nil)
	}()

	// Register a version gauge.
	versionGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: jasmine.ApplicationNamespace,
			Name:      "version_info",
			Help:      "Version of the application.",
		},
	)
	prometheus.MustRegister(versionGauge)
	versionValue, err := strconv.ParseFloat(version, 64)
	if err != nil {
		versionValue = 0.0
	}
	versionGauge.Set(versionValue)
}

func main() {
	log.Println("Coming online...")
	log.Printf("Version: %v\n", version)

	// Set up observability.
	initObservability()

	// Channel to be aware of an OS interrupt like Control-C.
	var waiter sync.WaitGroup
	waiter.Add(1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Read in the config file.
	configFile := flag.String("config", "config.yaml", "Path to the config file")
	flag.Parse()
	config, err := jasmine.ReadInConfig(*configFile)
	if err != nil {
		log.Printf("No config file specified or error reading in config: %v\n", err)
		log.Println("Running but not doing anything. Pretty boring, eh?")
	} else {
		log.Printf("Using '%s' as the jira server.\n", config.JiraServer)
		log.Printf("Loading %d queries for work.\n", len(config.Queries))

		// Do the work.
		jiraCoordinator, err := jasmine.NewJiraCoordinator(config)
		if err != nil {
			log.Fatalf("Error in creating jira coordinator: %v\n", err)
		}
		go jiraCoordinator.DoWork()
	}

	// Function and waiter to wait for the OS interrupt and do any clean-up.
	go func() {
		<-c
		fmt.Println("\r")
		log.Println("Interrupt captured.")
		waiter.Done()
	}()
	waiter.Wait()

	// Shut down the application.
	log.Println("Shutting down.")
}
