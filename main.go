package main

import (
	"flag"
	"log"
	"runtime/pprof"
	"sync"
	"time"

	cdns "github.com/niclabs/dnszeppelin"
	"os"
	"runtime"
)

var devName = flag.String("devName", "", "Device used to capture")
var pcapFile = flag.String("pcapFile", "", "Pcap filename to run")
// Filter is not using "(port 53)", as it will filter out fragmented udp 
// packets, instead, we filter by the ip protocol and check again in the
// application.
var batchSize = flag.Uint("batchSize", 10000000, "Minimun capacity of the cache array used to send data to clickhouse. Set close to the queries per minute received to prevent allocations")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var defraggerChannelSize = flag.Uint("defraggerChannelSize", 500, "Size of the channel to send packets to be defragged")
var defraggerChannelReturnSize = flag.Uint("defraggerChannelReturnSize", 500, "Size of the channel where the defragged packets are returned")
var filter = flag.String("filter", "((ip and (ip[9] == 6 or ip[9] == 17)) or (ip6 and (ip6[6] == 17 or ip6[6] == 6 or ip6[6] == 44)))", "BPF filter applied to the packet stream. If port is selected, the packets will not be defragged.")
var gcTime = flag.Uint("gcTime", 10, "Time in seconds to garbage collect the tcp assembly and ip defragmentation")
var influxdb = flag.String("influxdb", "localhost:8086", "Address of the Influx database to save the results")
var influxtoken = flag.String("influxtoken", "ratadns:ratadns", "InfluxDB token")
var influxorg = flag.String("influxorg", "", "InfluxDB organization")
var influxbucket = flag.String("influxbucket", "ratadns/weekly", "InfluxDB bucket")
var loggerFilename = flag.Bool("loggerFilename", false, "Show the file name and number of the logged string")
var memprofile = flag.String("memprofile", "", "write memory profile to file")
var packetHandlerCount = flag.Uint("packetHandlers", 1, "Number of routines used to handle received packets")
var packetChannelSize = flag.Uint("packetHandlerChannelSize", 100000, "Size of the packet handler channel")
var port = flag.Uint("port", 53, "Port selected to filter packets")
var resultChannelSize = flag.Uint("resultChannelSize", 100000, "Size of the result processor channel size")
var tcpAssemblyChannelSize = flag.Uint("tcpAssemblyChannelSize", 1000, "Size of the tcp assembler")
var tcpHandlerCount = flag.Uint("tcpHandlers", 1, "Number of routines used to handle tcp assembly")
var tcpResultChannelSize = flag.Uint("tcpResultChannelSize", 1000, "Size of the tcp result channel")
var wsize = flag.Uint("wsize", 60, "Size of processing window in seconds (default: 60)")

func checkFlags() {
	flag.Parse()
	if *loggerFilename {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags)
	}
	if *port > 65535 {
		log.Fatal("-port must be between 1 and 65535")
	}

	if *devName == "" && *pcapFile == "" {
		log.Fatal("-devName or -pcapFile is required")
	}

	if *devName != "" && *pcapFile != "" {
		log.Fatal("You must set only -devName or -pcapFile, and not both")
	}
}

func main() {
	checkFlags()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}
	resultChannel := make(chan cdns.DNSResult, *resultChannelSize)


	// Setup output routine
	exiting := make(chan bool)
	var wg sync.WaitGroup

	go collect(resultChannel, exiting, &wg, *wsize, *batchSize, *influxdb, *influxtoken, *influxorg, *influxbucket)

	// Setup mem profile
	if *memprofile != "" {
		go func() {
			time.Sleep(120 * time.Second)
				log.Println("Writing memory profile")
				f, err := os.Create(*memprofile)
				if err != nil {
					log.Fatal("could not create memory profile: ", err)
				}
				runtime.GC() // get up-to-date statistics
				if err := pprof.WriteHeapProfile(f); err != nil {
					log.Fatal("could not write memory profile: ", err)
				}
				f.Close()
		}()
	}

	// Start listening
	capturer := cdns.NewDNSCapturer(cdns.CaptureOptions{
		*devName,
		*pcapFile,
		*filter,
		uint16(*port),
		time.Duration(*gcTime) * time.Second,
		resultChannel,
		*packetHandlerCount,
		*packetChannelSize,
		*tcpHandlerCount,
		*tcpAssemblyChannelSize,
		*tcpResultChannelSize,
		*defraggerChannelSize,
		*defraggerChannelReturnSize,
		exiting,
	})
	capturer.Start()

	// Wait for the output to finish
	log.Println("Exiting")
	wg.Wait()
}
