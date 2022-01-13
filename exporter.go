package exporter

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/nftables"
	"github.com/prometheus/client_golang/prometheus"
)

var conn nftables.Conn

type NftablesCounterCollector struct {
	bytes   *prometheus.Desc
	packets *prometheus.Desc
}

func NewNftablesCounterCollector() (*NftablesCounterCollector, error) {
	tables, err := conn.ListTables()
	if err != nil {
		return nil, err
	}

	if len(tables) == 0 {
		return nil, fmt.Errorf("nftables: no table to export!")
	}

	return &NftablesCounterCollector{
		bytes: prometheus.NewDesc("nftables_counter_bytes",
			"Total number of bytes referenced by an nftables counter.",
			[]string{"family", "table", "counter"}, nil,
		),
		packets: prometheus.NewDesc("nftables_counter_packets",
			"Total number of packets referenced by an nftables counter.",
			[]string{"family", "table", "counter"}, nil,
		),
	}, nil
}

func (collector *NftablesCounterCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.bytes
	ch <- collector.packets
}

func getTableFamily(table *nftables.Table) string {
	switch table.Family {
	case nftables.TableFamilyINet:
		return "inet"
	case nftables.TableFamilyIPv4:
		return "ip"
	case nftables.TableFamilyIPv6:
		return "ip6"
	case nftables.TableFamilyARP:
		return "arp"
	case nftables.TableFamilyNetdev:
		return "netdev"
	case nftables.TableFamilyBridge:
		return "netdev"
	default:
		return "unknown"
	}
}

func cleanupName(name string) string {
	name = strings.ToLower(name)
	name = strings.Map(func(r rune) rune {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return '_'
		} else {
			return r
		}
	}, name)
	return name
}

func (collector *NftablesCounterCollector) collectTable(table *nftables.Table, ch chan<- prometheus.Metric) {
	objects, err := conn.GetObjects(table)
	if err != nil {
		log.Println(err)
		return
	}

	family := getTableFamily(table)
	tableName := cleanupName(table.Name)

	for i := range objects {
		counter, ok := objects[i].(*nftables.CounterObj)
		if !ok {
			// not a counter
			continue
		}

		counterName := cleanupName(counter.Name)
		ch <- prometheus.MustNewConstMetric(collector.bytes, prometheus.CounterValue, float64(counter.Bytes), family, tableName, counterName)
		ch <- prometheus.MustNewConstMetric(collector.packets, prometheus.CounterValue, float64(counter.Packets), family, tableName, counterName)
	}
}

func (collector *NftablesCounterCollector) Collect(ch chan<- prometheus.Metric) {
	tables, err := conn.ListTables()
	if err != nil {
		log.Println(err)
		return
	}

	for _, table := range tables {
		collector.collectTable(table, ch)
	}
}
