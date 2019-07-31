package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	namespace = "pactbroker"
)

var (
	// Version is set during build to the git describe version
	Version = "0.0.1"

	listenAddress = kingpin.Flag(
		"web.listen-address",
		"Address to listen on for web interface and telemetry.",
	).Default(getEnv("PB_EXPORTER_WEB_LISTEN_ADDRESS", ":9623")).String()
	metricPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default(getEnv("PB_EXPORTER_WEB_TELEMETRY_PATH", "/metrics")).String()
	dataSourceName = kingpin.Flag(
		"data-source-name",
		"Address of Pact Broker",
	).Default(getEnv("DATA_SOURCE_NAME", "http://localhost:9292")).String()

	netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

	pactBrokerUp = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Whether the last scrape of metrics from Pact Broker was able to connect to the server (1 for yes, 0 for no).",
		},
	)
	pactBrokerPacticipants = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "pacticipant_total",
			Help:      "The total number of pacticipants",
		},
	)
	pactBrokerPacts = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "pacticipant",
			Help:      "Current number of pacts",
		},
		[]string{"name"},
	)
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

// Check connection to Pact Broker
func checkConnection() (float64, error) {
	r, e := netClient.Get(*dataSourceName)
	if e != nil {
		return 0, e
	}
	if r.StatusCode >= 200 && r.StatusCode <= 299 {
		return 1, nil
	} else {
		return 0, nil
	}
}

// Part of structure pacticipants response
type pacticipants struct {
	Embedded struct {
		Pacticipants []struct {
			Name string `json:"name"`
		} `json:"pacticipants"`
	} `json:"_embedded"`
}

func getPacticipants() (pacticipants, error) {
	var p pacticipants

	r, e := netClient.Get(*dataSourceName + "/pacticipants")
	if e != nil {
		return p, e
	}
	defer r.Body.Close()

	body, e := ioutil.ReadAll(r.Body)
	if e != nil {
		return p, e
	}
	json.Unmarshal([]byte(body), &p)

	return p, nil
}

// Part of structure pacts response
type pacts struct {
	Links struct {
		Pbpacts []struct {
			Name string `json:"name"`
		} `json:"pb:pacts"`
	} `json:"_links"`
}

func getPacts(pacticipant string) (pacts, error) {
	var p pacts

	r, e := netClient.Get(*dataSourceName + "/pacts/provider/" + pacticipant)
	if e != nil {
		return p, e
	}
	defer r.Body.Close()

	body, e := ioutil.ReadAll(r.Body)
	if e != nil {
		return p, e
	}
	json.Unmarshal([]byte(body), &p)

	return p, nil
}

func main() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(fmt.Sprintf("pactbroker_exporter %s (built with %s)", Version, runtime.Version()))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	go func() {
		for {
			time.Sleep(2 * time.Second)

			c, e := checkConnection()
			pactBrokerUp.Set(c)
			if e != nil {
				log.Error(e)
				continue
			}

			p, e := getPacticipants()
			if e != nil {
				log.Error(e)
			}
			pactBrokerPacticipants.Set(float64(len(p.Embedded.Pacticipants)))

			for _, pacticipant := range p.Embedded.Pacticipants {
				pacts, e := getPacts(pacticipant.Name)
				if e != nil {
					log.Error(e)
				}
				pactBrokerPacts.WithLabelValues(pacticipant.Name).Set(float64(len(pacts.Links.Pbpacts)))
			}
		}
	}()

	// landingPage contains the HTML served at '/'.
	var landingPage = []byte(`<html><head><title>Pact Broker Exporter</title></head><body><h1>Pact Broker Exporter</h1><p><a href='` + *metricPath + `'>Metrics</a></p></body></html>`)

	log.Infoln("Starting pactbroker_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	http.Handle(*metricPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(landingPage)
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
