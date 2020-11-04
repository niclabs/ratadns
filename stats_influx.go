package main

import (
        "github.com/influxdata/influxdb-client-go/v2"
        "github.com/influxdata/influxdb-client-go/v2/api"
        cdns "github.com/niclabs/dnszeppelin"
	dns "github.com/miekg/dns"
	"strings"
        "time"
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

func InfluxAggAndStore(writeAPI api.WriteAPI, batch []cdns.DNSResult) error {
	defer writeAPI.Flush()

	type Q struct {
		qtype,ip,qname string
	}
	type R struct {
		rcode,ip,qname string
	}

	fields := make(map[string]int)
	sources := make(map[string]int)
	domains := make(map[string]int)
	info := make(map[Q]int)
	responses := make(map[R]int)

	if len(batch) == 0 {
		return nil
	}
	fields["TOTALQ"] = 0
	fields["TOTALR"] = 0
	fields["NOERROR"] = 0
	fields["NXDOMAIN"] = 0

	now :=  time.Now()

	for _,b := range batch {
		src := b.SrcIP.String()
		dst := b.DstIP.String()

		if b.DNS.Response  {
			fields["TOTALR"]++
			rc := dns.RcodeToString[b.DNS.Rcode]
			if b.DNS.Rcode == 0 {fields["NOERROR"]++}
			if b.DNS.Rcode == 3 {fields["NXDOMAIN"]++}
			name := ""
			if len(b.DNS.Question) > 0 {name=strings.ToLower(b.DNS.Question[0].Name)}
			responses[R{rc,dst,name}] = 1 + responses[R{rc,dst,name}]
		} else {
			fields["TOTALQ"] = 1 + fields["TOTALQ"]
			sources[src] = 1 + sources[src]
			for _,d := range b.DNS.Question {
				name := strings.ToLower(d.Name)
				domains[name] = 1 + domains[name]
				qt := dns.TypeToString[d.Qtype]
				fields[qt] = 1 + fields[qt]
				info[Q{qt,src,name}] = 1 + info[Q{qt,src,name}]
			}
		}
	}

	// Adding some stats
	fields["UNIQUERY"] = len(domains)

  // Store TNSM stats 
	go func() {
		for k,v := range fields {
			p := influxdb2.NewPoint("stat",
				map[string]string{"type" : k},
				map[string]interface{}{"freq" : v},
				now)
			writeAPI.WritePoint(p)
		}
	}()

	// Store also responses
	  go func() {
		for k,v := range responses {
			p := influxdb2.NewPoint("responses",
				map[string]string{
					"rcode" : k.rcode,
					"ip" : k.ip,
					"qname" : k.qname},
				map[string]interface{}{"freq" : v},
                                now)
			writeAPI.WritePoint(p)
		}
	}()

        // Store extra info
        go func() {
                for k,v := range info {
                        p := influxdb2.NewPoint("info",
                                map[string]string{
					"qtype" : k.qtype,
					"ip" : k.ip,
					"qname" : k.qname},
                                map[string]interface{}{"freq" : v},
                                now)
                        writeAPI.WritePoint(p)
                }
        }()

	return nil
}
