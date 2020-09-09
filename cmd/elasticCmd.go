package cmd

import (
	"errors"
	"fmt"
	"github.com/mono83/dogrelay/elastic"
	"github.com/mono83/dogrelay/udp"
	"github.com/mono83/xray"
	"github.com/mono83/xray/args"
	"github.com/spf13/cobra"
	"time"
)

var elasticLimitQueueCount, elasticLimitQueueSize int
var elasticUdpBufferSize int
var elasticUdpBind string
var elasticClientsCount int
var elasticDSN []string
var elasticIndexFormat string
var elasticEnsureTemplate bool

var elasticCmd = &cobra.Command{
	Use:   "elastic",
	Short: "Logstash replacement, that forwards data from UDP to ElasticSearch",
	RunE: func(cmd *cobra.Command, a []string) error {
		// Constructing byte dispatcher
		dis := udp.NewByteDispatcher(elasticLimitQueueCount, elasticLimitQueueSize)

		// Logging
		checkAndRunPrometheus()

		// Constructing elastic clients
		if elasticClientsCount < 1 {
			return errors.New("at least one client expected")
		}
		for i := 0; i < elasticClientsCount; i++ {
			cl, err := elastic.NewClient(elasticDSN, elasticIndexFormat, "", "")
			if err != nil {
				return err
			}
			if i == 0 && elasticEnsureTemplate {
				if err := cl.CreateIndexTemplate(); err != nil {
					return err
				}
			}

			go func(cl *elastic.Client) {
				for b := range dis.Channel() {
					err := cl.Write(b)
					fmt.Println(err)
				}
			}(cl)
		}

		// Exporting metrics
		go func() {
			log := xray.ROOT.Fork().WithMetricPrefix("bytequeue")
			for {
				time.Sleep(time.Second)
				count, size, dropByCount, dropBySize := dis.Stats()
				log.Gauge("count", int64(count))
				log.Gauge("size", int64(size))
				log.Gauge("drop", int64(dropByCount), args.Type("count"))
				log.Gauge("drop", int64(dropBySize), args.Type("size"))
			}
		}()

		// Starting UDP listening server
		if err := udp.StartServer(elasticUdpBind, elasticLimitQueueSize, dis.Publish); err != nil {
			return err
		}

		for {
			time.Sleep(time.Second)
		}
	},
}

func init() {
	elasticCmd.Flags().BoolVarP(&elasticEnsureTemplate, "ensure", "e", false, "If true, attempts to create index template")
	elasticCmd.Flags().IntVar(&elasticLimitQueueCount, "limit-count", 0, "Max items in delivery queue")
	elasticCmd.Flags().IntVar(&elasticLimitQueueSize, "limit-size", 0, "Max bytes in delivery queue")
	elasticCmd.Flags().IntVar(&elasticUdpBufferSize, "buffer", 8*4096, "UDP buffer size")
	elasticCmd.Flags().IntVarP(&elasticClientsCount, "count", "c", 1, "Count of clients (workers)")
	elasticCmd.Flags().StringVar(&elasticUdpBind, "bind", "", "UDP bind address")
	elasticCmd.Flags().StringVar(&elasticIndexFormat, "index", "logstash-2006.01.02", "Index time pattern according to Go time formatter")
	elasticCmd.Flags().StringVarP(&prometheusBind, "export-prometheus", "e", "", "Starts Prometheus exporter on given address, like :12345")
	elasticCmd.Flags().StringArrayVar(&elasticDSN, "elastic", []string{"http://localhost:9200"}, "ElasticSearch DSN, can be multiple")
}
