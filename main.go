package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	namespace = "pactbroker"
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

// Exporter structure
type Exporter struct {
	uri   string
	mutex sync.RWMutex
	fetch func(endpoint string) (io.ReadCloser, error)

	pactBrokerUp, pactBrokerPacticipants		prometheus.Gauge
	pactBrokerPacts                         *prometheus.GaugeVec
	totalScrapes                            prometheus.Counter
}

// NewExporter function
func NewExporter(uri string, timeout time.Duration) (*Exporter, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	var fetch func(endpoint string) (io.ReadCloser, error)
	switch u.Scheme {
	case "http", "https", "file":
		fetch = fetchHTTP(uri, timeout)
	default:
		return nil, fmt.Errorf("unsupported scheme: %q", u.Scheme)
	}

	return &Exporter{
		uri:   uri,
		fetch: fetch,
		pactBrokerUp: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "up",
				Help:      "Was the last scrape successful.",
			},
		),
		pactBrokerPacts: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "pacticipants",
				Help:      "Current number of pacts per pacticipant.",
		 	},
		 	[]string{"name"},
		),
		pactBrokerPacticipants: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "pacticipants_total",
				Help:      "The total number of pacticipants.",
			},
		),		
		totalScrapes: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "scrapes_total",
				Help:      "Current total scrapes.",
			},
		),
	}, nil
}

// Describe function of Exporter
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {}

// Collect function of Exporter
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() 
	defer e.mutex.Unlock()

	up := e.scrape(ch)

	ch <- prometheus.MustNewConstMetric(e.pactBrokerUp.Desc(), prometheus.GaugeValue, up)
}

type pacticipants struct {
	Embedded struct {
		Pacticipants []struct {
			Name string `json:"name"`
		} `json:"pacticipants"`
	} `json:"_embedded"`
}

type pacts struct {
	Links struct {
		Pbpacts []struct {
			Name string `json:"name"`
		} `json:"pb:pacts"`
	} `json:"_links"`
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) (up float64) {
	e.totalScrapes.Inc()

	var p pacticipants

	body, err := e.fetch("/pacticipants")
	if err != nil {
		log.Errorf("Can't scrape Pack: %v", err)
		return 0
	}
	defer body.Close()

	bodyAll, err := ioutil.ReadAll(body)
	if err != nil {
		return 0
	}

	_ = json.Unmarshal([]byte(bodyAll), &p)

	e.pactBrokerPacticipants.Set(float64(len(p.Embedded.Pacticipants)))

	for _, pacticipant := range p.Embedded.Pacticipants {
		var bodyPact io.ReadCloser
		var pacts pacts

		bodyPact, err = e.fetch("/pacts/provider/" + pacticipant.Name)
		if err != nil {
			log.Errorf("Can't scrape Pack: %v", err)
			return 0
		}
		defer bodyPact.Close()

		pactsAll, err := ioutil.ReadAll(bodyPact)
		if err != nil {
			return 0
		}
		_ = json.Unmarshal([]byte(pactsAll), &pacts)

		e.pactBrokerPacts.WithLabelValues(pacticipant.Name).Set(float64(len(pacts.Links.Pbpacts)))
	}

	return 1
}

func fetchHTTP(uri string, timeout time.Duration) func(endpoint string) (io.ReadCloser, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	return func(endpoint string) (io.ReadCloser, error) {
		resp, err := client.Get(uri + endpoint)
		if err != nil {
			return nil, err
		}
		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
		}
		return resp.Body, nil
	}
}

func main() {

	var (
		listenAddress   = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(getEnv("PB_EXPORTER_WEB_LISTEN_ADDRESS", ":9624")).String()
		metricsPath     = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default(getEnv("PB_EXPORTER_WEB_TELEMETRY_PATH", "/metrics")).String()
		uri 						= kingpin.Flag("pactbroker.uri", "URI of Pact Broker.").Default(getEnv("PB_EXPORTER_PACTBROKER_URI", "http://localhost:9292")).String()
		timeout   			= kingpin.Flag("pactbroker.timeout", "Scrape timeout").Default(getEnv("PB_EXPORTER_PACTBROKER_TIMEOUT", "5s")).Duration()
	)

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("pactbroker_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting pactbroker_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	exporter, err := NewExporter(*uri, *timeout)
	if err != nil {
		log.Fatal(err)
	}
	prometheus.MustRegister(exporter)
	prometheus.MustRegister(version.NewCollector("pactexporter"))

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><head><title>Pact Broker Exporter</title></head><body><h1>Pact Broker Exporter</h1><p><a href='` + *metricsPath + `'>Metrics</a></p></body></html>`))
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}