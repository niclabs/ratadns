package main

import (
        "github.com/influxdata/influxdb-client-go/v2"
        "github.com/influxdata/influxdb-client-go/v2/api"
        cdns "github.com/niclabs/dnszeppelin"
        "log"
        "sync"
        "time"
)


func collect(resultChannel chan cdns.DNSResult, exiting chan bool, wg *sync.WaitGroup, wsize, batchSize uint, influxdb, influxtoken, influxorg, influxbucket string){
        wg.Add(1)
        defer wg.Done()

        // Connect to InfluxDB
        client := influxdb2.NewClient(influxdb, influxtoken)
        defer client.Close()
        // Get non-blocking write client
        writeAPI := client.WriteAPI(influxorg, influxbucket)

        batch := make([]cdns.DNSResult, 0, batchSize)

        ticker := time.NewTicker(time.Duration(wsize) * time.Second)
  defer ticker.Stop()

        for {
                select {
                case data := <-resultChannel:
                        batch = append(batch, data)
                case <-ticker.C:
                        if err := aggAndStore(writeAPI,batch); err != nil {
                                log.Fatal("Error writing to DB:", err)
                                exiting <- true
                                return
                        } else {
                                batch = make([]cdns.DNSResult, 0, batchSize)
                        }
                case <-exiting:
                        return
                }
        }
}


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

	nqueries := make(map[uint16]int)
	domains := make(map[string]int)
	sources := make(map[string]int)
	responses := make(map[int]int)

	if len(batch) == 0 {
		return nil
	}

	for _,b := range batch {
		ip := b.SrcIP.String()
		sources[ip] = 1 + sources[ip]
		if b.DNS.Response  {
			responses[b.DNS.Rcode] = 1 + responses[b.DNS.Rcode]
		} else {
		for _,d := range b.DNS.Question {
			domains[d.Name] = 1 + domains[d.Name]
			nqueries[d.Qtype] = 1 + nqueries[d.Qtype]
			}
		}
	}
  // Store per qtypes
	go func() {
		for k,v := range nqueries {
			if k == 1 || k == 2 || k == 15 || k == 28 || k == 255 {
				log.Println("qType: ",k," frequency: ",v)
			}
		}
	}()

	// Store sources
  go func() {
		for k,v := range sources {
			log.Println("source: ",k," frequency: ",v)
		}
	}()

	// Store domain names
	go func() {
		for k,v := range domains {
			log.Println("name: ",k," frequency: ",v)
		}
	}()

	// Store responses
	  go func() {
    for k,v := range responses {
      if k == 0 || k == 3 {
        log.Println("qType: ",k," frequency: ",v)
      }
    }
  }()

	return nil
}
