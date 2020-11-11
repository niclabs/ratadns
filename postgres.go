package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	cdns "github.com/niclabs/dnszeppelin"
	"log"
	"sync"
	"time"
)


const (
	host	 = "localhost"
	dbport	 = 5432
)

func PGCollect(resultChannel chan cdns.DNSResult, exiting chan bool, wg *sync.WaitGroup, wsize, batchSize uint, dbname, user, pass string){

	wg.Add(1)
	defer wg.Done()

	// Connect to DB
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, dbport, user , pass, dbname)
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		log.Fatal("Error DB:",err)
	}
	defer db.Close()
	// check db
	err = db.Ping()
	if err != nil {
		log.Fatal("Error DB:",err)
	}

	batch := make([]cdns.DNSResult, 0, batchSize)

	ticker := time.NewTicker(time.Duration(wsize) * time.Second)
	defer ticker.Stop()

	for {
		select {
			case data := <-resultChannel:
				batch = append(batch, data)
			case <-ticker.C:
				if err := PGWrite(db,batch,300); err != nil { // this is Sparta!
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

