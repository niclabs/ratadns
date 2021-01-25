package main

import (
        "github.com/influxdata/influxdb-client-go/v2"
        "github.com/influxdata/influxdb-client-go/v2/api"
        cdns "github.com/niclabs/dnszeppelin"
        "log"
        "sync"
        "time"
)

type Database interface{
        createClient(d *database)
        createApi(d *database)
        InfluxCollect(resultChannel chan cdns.DNSResult, exiting chan bool, wg *sync.WaitGroup, wsize, batchSize uint, m *maps)
        InfluxStore(m *maps, batch []cdns.DNSResult) error
        StoreEachMap(mapa map[string]int, tipo1, tipo2 string , now time.Time)
}

type DefaultDB struct{
        influxdb string
        influxtoken string
        influxorg string
        influxbucket string
}

type database struct{
        c influxdb2.Client
        api api.WriteAPI
}

func (db DefaultDB) createClient(d *database){
        client := influxdb2.NewClient(db.influxdb, db.influxtoken)
        d.c = client
}

func (db DefaultDB) createApi(d *database){
        api := (d.c).WriteAPI(db.influxorg, db.influxbucket)
        d.api = api
}

func (d database) InfluxCollect(resultChannel chan cdns.DNSResult, exiting chan bool, wg *sync.WaitGroup, wsize, batchSize uint, m *maps){
        wg.Add(1)
        defer wg.Done()

        //Esto seria un metodo de estructura db

        // Connect to InfluxDB
        client := d.c
        defer client.Close()
        // Get non-blocking write client
        //writeAPI := d.api

        batch := make([]cdns.DNSResult, 0, batchSize)

        ticker := time.NewTicker(time.Duration(wsize) * time.Second)
  defer ticker.Stop()

        for {
                select {
                case data := <-resultChannel:
                        batch = append(batch, data)
                case <-ticker.C:
                        err := InfluxAgg(batch, m)
                        err1 := d.InfluxStore(m, batch)
                        if err != nil {
                                log.Fatal("Error writing to DB:", err)
                                exiting <- true
                                return
                        } else if err1 != nil{
                                log.Fatal("Error writing to DB:", err1)
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

