package main

import (
        "github.com/influxdata/influxdb-client-go/v2"
        cdns "github.com/niclabs/dnszeppelin"
        "log"
        "sync"
        "time"
)


func PGCollect(resultChannel chan cdns.DNSResult, exiting chan bool, wg *sync.WaitGroup, wsize, batchSize uint, influxdb, influxtoken, influxorg, influxbucket string){
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
                        if err := PGAggAndStore(writeAPI,batch); err != nil {
                                log.Fatal("Error writing to DB:", err)
                                exiting <- true
                                return
                        } else {
                                batch = make([]cdns.DNSResult, 0, batchSize)
                        }
                case <-exiting:
			exiting <- true
                        return
                }
        }
}

