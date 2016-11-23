package main

import (
	"flag"
	"fmt"
	"github.com/mono83/dogrelay/metrics"
	"github.com/mono83/dogrelay/udp"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	var size int
	var bind, influx string
	var compat bool
	var percString string

	flag.IntVar(&size, "size", 4096, "Packet size limit")
	flag.StringVar(&bind, "bind", "", "Listening port and address, for example localhost:8080")
	flag.StringVar(&influx, "influx", "", "InfluxDB target address and port to forward data")
	flag.StringVar(&percString, "percentiles", "95,98", "Percentiles to calculate, comma separated")
	flag.BoolVar(&compat, "compat", false, "StatsD compatible metrics mode. Will append .counter and .gauge for metrics")
	flag.Parse()
	if len(bind) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	var percentiles []int
	if len(percString) == 0 {
		fmt.Println("Percentiles not provided")
		os.Exit(4)

	}
	for _, v := range strings.Split(percString, ",") {
		iv, err := strconv.Atoi(v)
		if err != nil {
			fmt.Println(err)
			os.Exit(4)
		} else if iv < 50 || iv > 99 {
			fmt.Printf("Percentile must be in range [50, 99], but %d provided\n", iv)
			os.Exit(4)
		}

		percentiles = append(percentiles, iv)
		fmt.Printf("Will calculate %d-th percentile\n", iv)
	}

	if compat {
		fmt.Println("StatsD compatible outgoing metrics suffixes enabled")
	} else {
		fmt.Println("StatsD compatible outgoing metrics suffixes disabled")
	}

	buf := metrics.NewBuffer(percentiles, compat)
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
