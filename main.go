package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/nftables"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var conn nftables.Conn

type nftablesCollector struct {
	ipv4Traffic *prometheus.Desc
	ipv6Traffic *prometheus.Desc
}

func newNftablesCollector() *nftablesCollector {
	return &nftablesCollector{
		ipv4Traffic: prometheus.NewDesc("ipv4_traffic_bytes_counter",
			"Count how many bytes of IPv4 traffic this router has seen.",
			[]string{"direction"}, nil,
		),
		ipv6Traffic: prometheus.NewDesc("ipv6_traffic_bytes_counter",
			"Count how many bytes of IPv6 traffic this router has seen.",
			[]string{"direction"}, nil,
		),
	}
}

func (collector *nftablesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.ipv4Traffic
	ch <- collector.ipv6Traffic
}

func (collector *nftablesCollector) Collect(ch chan<- prometheus.Metric) {
	var table nftables.Table = nftables.Table{Name: "ip-counter", Family: nftables.TableFamilyINet}
	objects, err := conn.GetObjects(&table)
	if err != nil {
		log.Println(err)
		return
	}
	matches := 0
	for i := range objects {
		counter, ok := objects[i].(*nftables.CounterObj)
		if !ok {
			// not a counter
			continue
		}

		parts := strings.Split(counter.Name, "-")

		var desc *prometheus.Desc
		if parts[0] == "IPv4" {
			desc = collector.ipv4Traffic
		} else if parts[0] == "IPv6" {
			desc = collector.ipv6Traffic
		} else {
			continue
		}

		var direction string
		if parts[1] == "In" {
			direction = "incoming"
		} else if parts[1] == "Out" {
			direction = "outgoing"
		} else {
			continue
		}
		matches++

		bytes := prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(counter.Bytes), direction)
		ch <- bytes
	}

	if matches == 0 {
		log.Println("Collect: no object matched")
	}
}

func main() {
	tables, err := conn.ListTables()
	if err != nil {
		log.Fatal(err)
	}

	if len(tables) == 0 {
		log.Fatal("nftables: no table to show!")
	}

	registry := prometheus.NewPedanticRegistry()
	nftc := newNftablesCollector()
	registry.MustRegister(nftc)

	var opts promhttp.HandlerOpts
	opts.ErrorLog = log.Default()
	opts.ErrorHandling = promhttp.HTTPErrorOnError
	opts.MaxRequestsInFlight = 5
	opts.Timeout = 5 * time.Second
	opts.EnableOpenMetrics = true
	http.Handle("/metrics", promhttp.HandlerFor(registry, opts))
	log.Fatal(http.ListenAndServe(":9101", nil))
}
