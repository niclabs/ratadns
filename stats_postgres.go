package main

import (
        "database/sql"
	cdns "github.com/niclabs/dnszeppelin"
//	dns "github.com/miekg/dns"
	"log"
	"strings"
//	"time"
)

func PGWrite(db *sql.DB, batch []cdns.DNSResult, size int) error {

	l := len(batch)

	for i := 0; i < l; i = i + size {
		s := size
		if (i > l - size) { s = l-i }

		go func() {
			vs := make([]string,s)
			va := []interface{}{}

			tx,err := db.Begin()
			if err != nil {
				log.Fatal("Error TX Begin ", err)
			}
			for j := 0; j < size; j++ {
				vs[j] = "(?, ?, ?, ?, ?, ?, ?)"
				ts := batch[i + j].Timestamp
				src := batch[i + j].SrcIP.String()
				dst := batch[i + j].DstIP.String()
				qname := strings.ToLower(batch[i + j].DNS.Question[0].Name)
				isQuestion := ! batch[i + j].DNS.Response
				var code int
				if isQuestion {
					code = int(batch[i + j].DNS.Question[0].Qtype)
				} else {
					code = int(batch[i + j].DNS.Rcode)
				}
				pl := int(batch[i + j].PacketLength)
				va = append(va, ts, src, dst, qname, isQuestion, code, pl)
			}

			str := "INSERT INTO msg (time,src,dst,qname,iq,code,len) VALUES "
			str = str + strings.Join(vs,",")

			_, err = tx.Exec(str, va...)
			if err != nil {
				tx.Rollback()
				log.Fatal("Error Exec ", err)
			}
			err = tx.Commit()
			if err != nil {
				log.Fatal("Error db TX ",err)
			}
		} ()  // End go func()
	}
	return nil
}
