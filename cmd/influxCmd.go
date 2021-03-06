package cmd

import (
	"errors"
	"fmt"
	"github.com/mono83/dogrelay/metrics"
	"github.com/mono83/dogrelay/udp"
	v "github.com/mono83/validate"
	"github.com/mono83/xray"
	"github.com/mono83/xray/args"
	"github.com/spf13/cobra"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var influxCmdPktSize int
var influxCmdBind, influxCmdInfluxHost, influxCmdPercString string
var influxCmdCompatMode bool

var influxCmd = &cobra.Command{
	Use:   "statsd-influx",
	Short: "Starts buffered UDP reader, that flushes data to InfluxDb",
	RunE: func(cmd *cobra.Command, a []string) error {
		if err := v.All(
			v.WithMessage(v.StringNotWhitespace(influxCmdPercString), "Percentiles not provided"),
			v.WithMessage(v.StringNotWhitespace(influxCmdInfluxHost), "InfluxDB host not provided"),
			v.WithMessage(v.StringNotWhitespace(influxCmdBind), "Binding address not provided"),
		); err != nil {
			return err
		}

		// Parsing percentiles
		var percentiles []int
		for _, v := range strings.Split(influxCmdPercString, ",") {
			iv, err := strconv.Atoi(v)
			if err != nil {
				xray.BOOT.Error("Error parsing percentiles - :err", args.Error{Err: err})
				return err
			} else if iv < 50 || iv > 99 {
				xray.BOOT.Error(
					"Percentile must be in range [50, 99], but :value provided",
					args.Int{N: "value", V: iv},
				)
				return errors.New("invalid percentile value")
			}

			percentiles = append(percentiles, iv)
			xray.BOOT.Info("Will calculate :value -th percentile\n", args.Int{N: "value", V: iv})
		}

		if influxCmdCompatMode {
			xray.BOOT.Info("StatsD compatible outgoing metrics suffixes enabled")
		} else {
			xray.BOOT.Info("StatsD compatible outgoing metrics suffixes disabled")
		}

		buf := metrics.NewBuffer(percentiles, influxCmdCompatMode)
		err := udp.StartMetricsServer(influxCmdBind, influxCmdPktSize, buf.Add)
		if err != nil {
			xray.BOOT.Error("Error starting UDP server - :err", args.Error{Err: err})
			return err
		}
		xray.BOOT.Info(
			"Listening incoming UDP on :addr with packet size below :count bytes",
			args.String{N: "addr", V: influxCmdBind},
			args.Count(influxCmdPktSize),
		)

		var inf *udp.InfluxDBSender
		if len(influxCmdInfluxHost) > 0 {
			inf, err = udp.NewInfluxDBSender(influxCmdInfluxHost)
			if err != nil {
				xray.BOOT.Error("Error starting InfluxDB  - :err", args.Error{Err: err})
				return err
			}
			xray.BOOT.Info(
				"Forwarding data to InfluxDB on :addr",
				args.String{N: "addr", V: influxCmdBind},
			)
		}

		name, err := os.Hostname()
		if err != nil {
			name = "unknown"
		} else {
			regex := regexp.MustCompile("[^\\w]")
			name = regex.ReplaceAllString(name, "")
			xray.BOOT.Info("Effective hostname is :name", args.Name(name))
		}

		params := []string{"hostname=" + name}
		checkAndRunPrometheus()

		for {
			time.Sleep(10 * time.Second)
			before := time.Now()
			toSend, rawCount, aggCount := buf.Flush(10)
			if inf == nil {
				fmt.Println()
				for _, e := range toSend {
					fmt.Println(e.Value, "\t", e.Key())
				}
			} else {
				for _, e := range toSend {
					inf.Send(e)
				}
			}

			// Adding system metric to inform about time spent to aggregate data
			buf.Add(metrics.Event{
				EventType: metrics.TypeGauge,
				Value:     time.Now().Sub(before).Nanoseconds(),
				Metric:    "dogrelay.pause",
				Params:    params,
			})

			// Incoming metrics count
			buf.Add(metrics.Event{
				EventType: metrics.TypeIncrement,
				Value:     int64(rawCount),
				Metric:    "dogrelay.in",
				Params:    params,
			})

			// Outgoing (aggregated) metrics count
			buf.Add(metrics.Event{
				EventType: metrics.TypeIncrement,
				Value:     int64(aggCount),
				Metric:    "dogrelay.out",
				Params:    params,
			})
		}
	},
}

func init() {
	influxCmd.Flags().IntVar(&influxCmdPktSize, "size", 4096, "Packet size limit")
	influxCmd.Flags().StringVar(&influxCmdBind, "bind", "", "Listening port and address, for example localhost:8080")
	influxCmd.Flags().StringVar(&influxCmdInfluxHost, "influx", "", "InfluxDB target address and port to forward data")
	influxCmd.Flags().StringVar(&influxCmdPercString, "percentiles", "95,98", "Percentiles to calculate, comma separated")
	influxCmd.Flags().BoolVar(&influxCmdCompatMode, "compat", false, "StatsD compatible metrics mode. Will append .counter and .gauge for metrics")
	influxCmd.Flags().StringVarP(&prometheusBind, "export-prometheus", "e", "", "Starts Prometheus exporter on given address, like :12345")
}
