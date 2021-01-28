package main

import (
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	dns "github.com/miekg/dns"
	cdns "github.com/niclabs/dnszeppelin"
)

/*
From:
Detecting Anomalies at a TLD Name Server Based on DNS Traffic Predictions
Diego Madariaga, Javier Madariaga, Martı́n Panza, Javier Bustos-Jiménez,
and Benjamin Bustos
IEEE Transactions on Network and Service Management (IEEE TNSM) Journal
Given the aforementioned, our proposed AD-BoP method
is focused on the following nine DNS traffic features:
• Number of DNS queries of types A (1), AAAA (2), NS
(3), MX (4), and ANY (5).
• Number of unique queried domains (6).
• Number of DNS response packets with codes NXDOMAIN
(7) and NOERROR (8).
• Total number of DNS packets (9).
*/

type maps struct {
	fields    map[string]int
	sources   map[string]int
	domains   map[string]int
	responses map[int]int
	filter    map[string][]float64
}

func InfluxAgg(batch []cdns.DNSResult, m *maps) error {

	fields := m.fields
	sources := m.sources
	domains := m.domains
	responses := m.responses

	if len(batch) == 0 {
		return nil
	}

	fields["TOTALQ"] = 0
	fields["TOTALR"] = 0

	for _, b := range batch {
		ip := b.SrcIP.String()
		if b.DNS.Response {
			fields["TOTALR"] = 1 + fields["TOTALR"]
			responses[b.DNS.Rcode] = 1 + responses[b.DNS.Rcode]
		} else {
			fields["TOTALQ"] = 1 + fields["TOTALQ"]
			for _, d := range b.DNS.Question {
				qt := dns.TypeToString[d.Qtype]
				name := strings.ToLower(d.Name)
				domains[name] = 1 + domains[name]
				sources[ip] = 1 + sources[ip]
				fields[qt] = 1 + fields[qt]
			}
		}
	}

	// Adding some stats
	fields["NOERROR"] = responses[0]
	fields["NXDOMAIN"] = responses[3]
	fields["UNIQUERY"] = len(domains)

	return nil
}

func emafilter(m *maps, number int, ttype string) error {

	if m.filter["DATA"+ttype] == nil {
		var array []float64
		m.filter["DATA"+ttype] = array
	}

	if len(m.filter["DATA"+ttype]) > number {
		step := Emastep(float64(number), float64(m.fields[ttype]), float64(m.fields["TREND"+ttype]))
		m.fields["TREND"+ttype] = int(step)
	} else if len(m.filter["DATA"+ttype]) == number {
		m.filter["DATA"+ttype] = append(m.filter["DATA"+ttype], float64(m.fields[ttype]))
		filtered := Ema(number, m.filter["DATA"+ttype])
		//si se quiere registrar los primeros "number" estimaciones:
		//m.filter["FIRSTTREND"+ttype] = filtered
		m.fields["TREND"+ttype] = int(filtered[len(filtered)-1])
	} else {
		m.filter["DATA"+ttype] = append(m.filter["DATA"+ttype], float64(m.fields[ttype]))
	}

	return nil
}

func (d database) InfluxStore(m *maps, batch []cdns.DNSResult) error {
	if len(batch) == 0 {
		return nil
	}

	now := time.Now()
	defer d.api.Flush()
	// Store TNSM stats
	go d.StoreEachMap(m.fields, "stat", "type", now)
	// Store also sources
	go d.StoreEachMap(m.sources, "source", "ip", now)
	// Store domain names
	go d.StoreEachMap(m.domains, "domain", "qname", now)

	return nil
}

func (d database) StoreEachMap(mapa map[string]int, metric, field string, now time.Time) {
	for k, v := range mapa {
		p := influxdb2.NewPoint(metric,
			map[string]string{field: k},
			map[string]interface{}{"freq": v},
			now)
		d.api.WritePoint(p)
	}
}
