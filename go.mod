module github.com/niclabs/ratadns

go 1.13

require (
	github.com/golang/mock v1.4.4
	github.com/influxdata/influxdb-client-go v1.4.0
	github.com/influxdata/influxdb-client-go/v2 v2.1.0
	github.com/lib/pq v1.8.0
	github.com/miekg/dns v1.1.33
	github.com/niclabs/dnszeppelin v1.1.0
	github.com/stretchr/testify v1.4.0
	github.com/vektra/mockery v1.1.2 // indirect
	golang.org/x/tools v0.1.0 // indirect
)

replace github.com/miekg/dns v1.1.33 => github.com/niclabs/dns v1.1.33
