package cmd

import (
	"github.com/mono83/xray"
	"github.com/mono83/xray/args"
	"github.com/mono83/xray/mon"
	"github.com/mono83/xray/out/prometheus"
	"net/http"
	"time"
)

var prometheusBind string

// checkAndRunPrometheus checks, if Prometheus exporter is enabled and if true
// runs it with health monitor
func checkAndRunPrometheus() {
	if len(prometheusBind) > 0 {
		// Starting self metrics
		mon.StartHealthMonitor(xray.ROOT)

		// Starting Prometheus exporter
		exporter := prometheus.NewExporter(
			func(in []xray.Arg) (out []xray.Arg) {
				for _, a := range in {
					switch a.Name() {
					case "host", "name", "type":
						out = append(out, a)
					}
				}
				return
			},
			nil,
			time.Millisecond,
			time.Millisecond*5,
			time.Millisecond*10,
			time.Millisecond*50,
			time.Millisecond*100,
			time.Millisecond*250,
			time.Second,
		)
		xray.ROOT.On(exporter.Handle)
		go func() {
			xray.BOOT.Info("Starting self diagnostics Prometheus exporter at :addr", args.Addr(prometheusBind))
			if err := http.ListenAndServe(prometheusBind, exporter); err != nil {
				xray.BOOT.Error("Error starting Prometheus exporter - :err", args.Error{Err: err})
			}
		}()
	}
}
