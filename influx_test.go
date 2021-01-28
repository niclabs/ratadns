package main

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	cdns "github.com/niclabs/dnszeppelin"
	"sync"
	dns "github.com/miekg/dns"

	"time"
	"net"
	"strconv"
)

// Function that creates a batch with 8 elements
func createBatch() []cdns.DNSResult{
	rChannel := make(chan cdns.DNSResult, 8)
	batch := make([]cdns.DNSResult, 0)
	timestamp := time.Now()
	ip := net.IP{}
	types := []uint16{dns.TypeA, dns.TypeSOA, dns.TypeIXFR, dns.TypeAXFR, dns.TypeA, dns.TypeA, dns.TypeIXFR}
	for i := 0; i < 7; i++{
		data := new(dns.Msg)
		data.SetQuestion("example"+strconv.Itoa(i)+".com.", types[i])
		dnsresult := cdns.DNSResult{timestamp, *data, uint8(0), ip, ip, "tcp", uint16(0)}
		rChannel <- dnsresult
		d := <-rChannel
    	batch = append(batch, d)
	}
	datar := new(dns.Msg)
	datar.Response = true
	dnsresult := cdns.DNSResult{timestamp, *datar, uint8(0), ip, ip, "tcp", uint16(0)}
	rChannel <- dnsresult
	d := <-rChannel
    batch = append(batch, d)

	return batch
}

// Tests InfluxAgg with empty batch
func TestInfluxAgg(t *testing.T){
	fields := make(map[string]int)
	sources := make(map[string]int)
	domains := make(map[string]int)
	responses := make(map[int]int)
	testmap := maps{fields, sources, domains, responses}	

	batch := make([]cdns.DNSResult, 0)
	got := InfluxAgg(batch, &testmap)
	if got != nil{
		t.Errorf("InfluxAgg(batch, &testmap) = %d; want nil", got)
	}

	if testmap.fields["TOTALR"] != 0{
		t.Errorf("fields[TOTALR] = %d; want 0", testmap.fields["TOTALR"])
	}

	if testmap.fields["TOTALQ"] != 0{
		t.Errorf("fields[TOTALQ] = %d; want 0", testmap.fields["TOTALQ"])
	}

	if testmap.fields["NOERROR"] != testmap.responses[0]{
		t.Errorf("fields[NOERROR] = %d; want responses[0]", testmap.fields["NOERROR"])
	}

	if testmap.fields["NXDOMAIN"] != testmap.responses[3]{
		t.Errorf("fields[NXDOMAIN] = %d; want responses[3]", testmap.fields["NXDOMAIN"])
	}

	if testmap.fields["UNIQUERY"] != 0{
		t.Errorf("fields[UNIQUERY] = %d; want 0", testmap.fields["UNIQUERY"])
	}
}

// Tests InfluxAgg with a batch created with createBatch
func TestInfluxAgg1(t *testing.T){
    batch := createBatch()

	fields := make(map[string]int)
	sources := make(map[string]int)
	domains := make(map[string]int)
	responses := make(map[int]int)
	testmap := maps{fields, sources, domains, responses}	

	got := InfluxAgg(batch, &testmap)
	if got != nil{
		t.Errorf("InfluxAgg(batch, &testmap) = %d; want nil", got)
	}

	if testmap.fields["TOTALQ"] != 7{
		t.Errorf("fields[TOTALQ] = %d; want 7", testmap.fields["TOTALQ"])
	}

	if testmap.fields["TOTALR"] != 1{
		t.Errorf("fields[TOTALR] = %d; want 1", testmap.fields["TOTALR"])
	}

	for i:=0; i < 4; i++{
		if testmap.domains["example"+strconv.Itoa(i)+".com."] != 1{
			t.Errorf("domains[examplei.com.] = %d; want 1", testmap.domains["example"+strconv.Itoa(i)+".com."])
		}
	}

	if testmap.sources["<nil>"] != 7{
		t.Errorf("sources[<nil>] = %d; want 7", testmap.sources["<nil>"])
	}

	if testmap.fields["A"] != 3{
		t.Errorf("fields[A] = %d; want 3", testmap.fields["A"])
	}

	if testmap.fields["SOA"] != 1{
		t.Errorf("fields[SOA] = %d; want 1", testmap.fields["SOA"])
	}

	if testmap.fields["IXFR"] != 2{
		t.Errorf("fields[IXFR] = %d; want 2", testmap.fields["IXFR"])
	}

	if testmap.fields["AXFR"] != 1{
		t.Errorf("fields[AXFR] = %d; want 1", testmap.fields["AXFR"])
	}
}

// Tests InfluxStore with empty batch
func TestInfluxStore(t *testing.T) {
	fields := make(map[string]int)
	sources := make(map[string]int)
	domains := make(map[string]int)
	responses := make(map[int]int)
	testmap := maps{fields, sources, domains, responses}
	batch := make([]cdns.DNSResult, 0)	

	ctrl := gomock.NewController(t)
  
	defer ctrl.Finish()
  
	m := NewMockDatabase(ctrl)
  
	m.
	  EXPECT().
	  InfluxStore(&testmap, batch).
	  Return(nil)
  
	m.InfluxStore(&testmap, batch)
}

// Tests InfluxStore with a batch created with createBatch
func TestInfluxStore1(t *testing.T) {
	fields := make(map[string]int)
	sources := make(map[string]int)
	domains := make(map[string]int)
	responses := make(map[int]int)
	testmap := maps{fields, sources, domains, responses}

	batch := createBatch()

	ctrl := gomock.NewController(t)
  
	defer ctrl.Finish()
  
	m := NewMockDatabase(ctrl)
  
	m.
	  EXPECT().
	  InfluxStore(&testmap, batch).
	  Return(nil)
  
	m.InfluxStore(&testmap, batch)
}

// Tests InfluxStore with an empty resultChannel 
func TestInfluxCollect(t *testing.T) {
	fields := make(map[string]int)
	sources := make(map[string]int)
	domains := make(map[string]int)
	responses := make(map[int]int)
	testmap := maps{fields, sources, domains, responses}
	resultChannel := make(chan cdns.DNSResult, *resultChannelSize)
	exiting := make(chan bool)
	var wg sync.WaitGroup	
	ctrl := gomock.NewController(t)
	
	defer ctrl.Finish()
  
	m := NewMockDatabase(ctrl)

	m.
	  EXPECT().
	  InfluxCollect(resultChannel, exiting, &wg, *wsize, *batchSize, &testmap)
  
	m.InfluxCollect(resultChannel, exiting, &wg, *wsize, *batchSize, &testmap)
}

// Tests InfluxStore with a non empty resultChannel
func TestInfluxCollect1(t *testing.T) {
	fields := make(map[string]int)
	sources := make(map[string]int)
	domains := make(map[string]int)
	responses := make(map[int]int)
	testmap := maps{fields, sources, domains, responses}
	resultChannel := make(chan cdns.DNSResult, *resultChannelSize)
	exiting := make(chan bool)
	var wg sync.WaitGroup	
	ctrl := gomock.NewController(t)

	timestamp := time.Now()
	data := new(dns.Msg)
	data.SetQuestion("example.com.", dns.TypeA)
	ip := net.IP{}
	dnsresult := cdns.DNSResult{timestamp, *data, uint8(0), ip, ip, "tcp", uint16(0)}
	resultChannel <- dnsresult
	
	defer ctrl.Finish()
  
	m := NewMockDatabase(ctrl)
  
	m.
	  EXPECT().
	  InfluxCollect(resultChannel, exiting, &wg, *wsize, *batchSize, &testmap)
  
	m.InfluxCollect(resultChannel, exiting, &wg, *wsize, *batchSize, &testmap)
}

// Tests StoreEachMap with the map fields
func TestStoreEachMap(t *testing.T) {
	fields := make(map[string]int)
	sources := make(map[string]int)
	domains := make(map[string]int)
	responses := make(map[int]int)
	testmap := maps{fields, sources, domains, responses}
	now := time.Now()
	ctrl := gomock.NewController(t)
  
	defer ctrl.Finish()
  
	m := NewMockDatabase(ctrl)
  
	m.
	  EXPECT().
	  StoreEachMap(testmap.fields, "stat", "type",  now)
  
	m.StoreEachMap(testmap.fields, "stat", "type",  now)
}

