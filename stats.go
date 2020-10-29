package main

import (
        "github.com/influxdata/influxdb-client-go/v2"
        "github.com/influxdata/influxdb-client-go/v2/api"
        cdns "github.com/niclabs/dnszeppelin"
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

func aggAndStore(writeAPI api.WriteAPI, batch []cdns.DNSResult) error {
	defer writeAPI.Flush()

	fields := make(map[string]int)
	domains := make(map[string]int)
	sources := make(map[string]int)
	responses := make(map[int]int)
	qtype := map[uint16]string {
		1 : "A",
		2 : "NS",
		15: "MX",
		28: "AAAA",
		255 : "ANY",
		}


	if len(batch) == 0 {
		return nil
	}
	fields["TOTALQ"] = 0
	fields["TOTALR"] = 0

	now :=  time.Now()

	for _,b := range batch {
		ip := b.SrcIP.String()
		if b.DNS.Response  {
			fields["TOTALR"] = 1 + fields["TOTALR"]
			responses[b.DNS.Rcode] = 1 + responses[b.DNS.Rcode]
		} else {
			fields["TOTALQ"] = 1 + fields["TOTALQ"]
			sources[ip] = 1 + sources[ip]
			for _,d := range b.DNS.Question {
				name := strings.ToLower(d.Name)
				domains[name] = 1 + domains[name]
				qt := d.Qtype
				if qt == 1 || qt == 2 || qt == 15 || qt == 28 || qt == 255 {
					fields[qtype[qt]] = 1 + fields[qtype[qt]]
				}
			}
		}
	}

	// Adding some stats
	for _,v := range qtype {
		if fields[v] == 0  { fields[v] = 0 }
	}
	fields["NOERROR"] = responses[0]
	fields["NXDOMAIN"] = responses[3]
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

	// Store also sources
  go func() {
		for k,v := range sources {
			p := influxdb2.NewPoint("source",
				map[string]string{"ip" : k},
				map[string]interface{}{"freq" : v},
                                now)
			writeAPI.WritePoint(p)
		}
	}()

	// Store domain names
	go func() {
		for k,v := range domains {
			p := influxdb2.NewPoint("domain",
				map[string]string{"qname" : k},
				map[string]interface{}{"freq" : v},
                                now)
			writeAPI.WritePoint(p)
		}
	}()

	return nil
}
