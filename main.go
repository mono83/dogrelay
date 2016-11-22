package main

import (
	"flag"
	"fmt"
	"github.com/mono83/dogrelay/metrics"
	"github.com/mono83/dogrelay/udp"
	"os"
	"time"
)

func main() {
	var size int
	var bind, influx string

	flag.IntVar(&size, "size", 2048, "Packet size limit")
	flag.StringVar(&bind, "bind", "", "Listening port and addressm, for example localhost:8080")
	flag.StringVar(&influx, "influx", "", "InfluxDB target address and port to forward data")
	flag.Parse()
	if len(bind) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	buf := metrics.NewBuffer([]int{90, 95, 98})
	err := udp.StartServer(bind, size, buf.Add)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	} else {
		fmt.Printf("Listening incoming UDP on %s with packet size below %d bytes\n", bind, size)
	}

	var inf *udp.InfluxDBSender
	if len(influx) > 0 {
		inf, err = udp.NewInfluxDBSender(influx)
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		} else {
			fmt.Printf("Forwarding data to InfluxDB on %s\n", influx)
		}
	}

	for {
		time.Sleep(10 * time.Second)
		if inf == nil {
			fmt.Println()
			for _, e := range buf.Flush() {
				fmt.Println(e.Value, "\t", e.Key())
			}
		} else {
			for _, e := range buf.Flush() {
				inf.Send(e)
			}
		}
	}
}
