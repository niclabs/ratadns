package main

import (
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	cdns "github.com/niclabs/dnszeppelin"
	"log"
	"sync"
	"time"
)


func dosomething(resultChannel chan cdns.DNSResult, exiting chan bool, wg *sync.WaitGroup, wsize, batchSize uint, influxdb, influxtoken, influxorg, influxbucket string){
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
			if err := sendPoints(writeAPI,batch); err != nil {
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

func sendPoints(writeAPI api.WriteAPI, batch []cdns.DNSResult) error {
	defer writeAPI.Flush()

	//features := make(map[string]int)
	domains := make(map[string]int)
	sources := make(map[string]int)

	if len(batch) == 0 {
		return nil
	}

	for _,b := range batch {
		ip := b.SrcIP.String()
		sources[ip] = 1 + sources[ip]

		for _,d := range b.DNS.Question {
			domains[d.Name] = 1 + domains[d.Name]
		}
	}

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
	return nil
}
