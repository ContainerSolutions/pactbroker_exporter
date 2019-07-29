package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Flags
var (
	listenAddress  string
	metricPath     string
	dataSourceName string
)

// Metric name parts.
const (
	// Namespace for all metrics.
	namespace = "pactbroker"
)

// Metrics variables
var (
	pactBrokerUp = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "up",
		Help:      "Whether the last scrape of metrics from Pact Broker was able to connect to the server (1 for yes, 0 for no).",
	})
	pactBrokerPacticipants = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "pacticipants_total",
		Help:      "The total number of pacticipants",
	})
)

func init() {
	flag.StringVar(&listenAddress, "web.listen-address", getEnv("PB_EXPORTER_WEB_LISTEN_ADDRESS", ":9623"), "Address to listen on for web interface and telemetry.")
	flag.StringVar(&metricPath, "web.telemetry-path", getEnv("PB_EXPORTER_WEB_TELEMETRY_PATH", "/metrics"), "Path under which to expose metrics.")
	flag.StringVar(&dataSourceName, "data-source-name", getEnv("DATA_SOURCE_NAME", "http://localhost:9292"), "Address of Pact Broker")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func checkPacticipants() (int, error) {
	r, e := http.Get(dataSourceName + "/pacticipants")
	if e != nil {
		return 0, e
	}
	defer r.Body.Close()

	var result map[string]interface{}

	body, e := ioutil.ReadAll(r.Body)
	if e != nil {
		return 0, e
	}

	json.Unmarshal([]byte(body), &result)

	return len(result["pacticipants"].([]interface{})), nil
}

func main() {
	flag.Parse()

	pactBrokerUp.Set(1)
	pacticipants, e := checkPacticipants()
	if e != nil {
		log.Fatal(e)
	}

	for i := 0; i < pacticipants; i++ {
		pactBrokerPacticipants.Inc()
	}

	http.Handle(metricPath, promhttp.Handler())
	log.Print("Starting Server: ", listenAddress)
	http.ListenAndServe(listenAddress, nil)

}
