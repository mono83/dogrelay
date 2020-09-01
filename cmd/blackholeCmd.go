package cmd

import (
	"github.com/mono83/dogrelay/udp"
	v "github.com/mono83/validate"
	"github.com/mono83/xray"
	"github.com/mono83/xray/args"
	"github.com/spf13/cobra"
	"time"
)

var blackholePktSize int
var blackholeBind string

var blackholeCmd = &cobra.Command{
	Use:     "blackhole",
	Aliases: []string{"bh"},
	Short:   "Starts UDP listener, that only receives packets",
	RunE: func(cmd *cobra.Command, a []string) error {
		if err := v.All(
			v.WithMessage(v.StringNotWhitespace(blackholeBind), "Binding address not provided"),
		); err != nil {
			return err
		}

		checkAndRunPrometheus()
		xray.BOOT.Info("Starting UDP listener for blackhole at :addr", args.Addr(blackholeBind))
		if err := udp.StartServer(blackholeBind, blackholePktSize, func(bts []byte) {}); err != nil {
			return err
		}

		for {
			time.Sleep(time.Second)
		}

		return nil
	},
}

func init() {
	blackholeCmd.Flags().IntVar(&influxCmdPktSize, "size", 8192, "Packet size limit")
	blackholeCmd.Flags().StringVar(&blackholeBind, "bind", "", "Listening port and address, for example localhost:8080")
	blackholeCmd.Flags().StringVarP(&prometheusBind, "export-prometheus", "e", "", "Starts Prometheus exporter on given address, like :12345")
}
