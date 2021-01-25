package main

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	cdns "github.com/niclabs/dnszeppelin"
	"sync"
)



func TestInfluxAgg(t *testing.T){

	fields := make(map[string]int)
	sources := make(map[string]int)
	domains := make(map[string]int)
	responses := make(map[int]int)
	mockmap := maps{fields, sources, domains, responses}	

	batch := make([]cdns.DNSResult, 0)
	got := InfluxAgg(batch, &mockmap)
	if got != nil{
		t.Errorf("InfluxAgg(batch, &mockmap) = %d; want nil", got)
	}

	batch1 := make([]cdns.DNSResult, 1)
	got1 := InfluxAgg(batch1, &mockmap)

	if got1 != nil{
		t.Errorf("InfluxAgg(batch1, &mockmap) = %d; want nil", got)
	}

	if mockmap.fields["TOTALR"] != 0{
		t.Errorf("fields[TOTALR] = %d; want 0", mockmap.fields["TOTALR"])
	}

	if mockmap.fields["TOTALQ"] != 1{
		t.Errorf("fields[TOTALQ] = %d; want q", mockmap.fields["TOTALQ"])
	}

	if mockmap.fields["NOERROR"] != mockmap.responses[0]{
		t.Errorf("fields[NOERROR] = %d; want responses[0]", mockmap.fields["NOERROR"])
	}

	if mockmap.fields["NXDOMAIN"] != mockmap.responses[3]{
		t.Errorf("fields[NOERROR] = %d; want responses[3]", mockmap.fields["NXDOMAIN"])
	}

	if mockmap.fields["UNIQUERY"] != 0{
		t.Errorf("fields[NOERROR] = %d; want 0", mockmap.fields["UNIQUERY"])
	}
}


func TestInfluxStore(t *testing.T) {
	fields := make(map[string]int)
	sources := make(map[string]int)
	domains := make(map[string]int)
	responses := make(map[int]int)
	mockmap := maps{fields, sources, domains, responses}
	batch := make([]cdns.DNSResult, 0)	
	ctrl := gomock.NewController(t)
  
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
  
	m := NewMockDatabase(ctrl)
  
	// Asserts that the first and only call to Bar() is passed 99.
	// Anything else will fail.
	m.
	  EXPECT().
	  InfluxStore(mockmap, batch).
	  Return(nil)
  
  }

  func TestInfluxCollect(t *testing.T) {
	fields := make(map[string]int)
	sources := make(map[string]int)
	domains := make(map[string]int)
	responses := make(map[int]int)
	mockmap := maps{fields, sources, domains, responses}
	batch := make([]cdns.DNSResult, 0)
	batchsize := 0
	resultChannel := make(chan cdns.DNSResult, *resultChannelSize)
	exiting := make(chan bool)
	var wg sync.WaitGroup	
	wsize := 60
	ctrl := gomock.NewController(t)
	
  
	// Assert that Bar() is invoked.
	defer ctrl.Finish()
  
	m := NewMockDatabase(ctrl)
  
	// Asserts that the first and only call to Bar() is passed 99.
	// Anything else will fail.
	m.
	  EXPECT().
	  InfluxCollect(resultChannel, exiting, &wg, wsize, batchSize, &mockmap)
  
  }