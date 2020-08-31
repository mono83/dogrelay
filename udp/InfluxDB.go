package udp

import (
	"bytes"
	"github.com/mono83/dogrelay/metrics"
	"github.com/mono83/udpwriter"
	"github.com/mono83/xray"
	"io"
	"strconv"
	"time"
)

// InfluxDBSender is adapter, used to send metrics events to InfluxDB
type InfluxDBSender struct {
	writer io.Writer
	log    xray.Ray
}

// NewInfluxDBSender builds and returns new adapter for InfluxDB
func NewInfluxDBSender(addr string) (*InfluxDBSender, error) {
	var err error
	snd := new(InfluxDBSender)
	snd.writer, err = udpwriter.NewS(addr)
	if err != nil {
		return nil, err
	}

	snd.log = xray.ROOT.Fork().WithLogger("indfluxdb-sender").WithMetricPrefix("influxdb")

	return snd, nil
}

// Send sends packet to InfluxDB
func (i *InfluxDBSender) Send(e metrics.Event) {
	// Converting to influxdb format
	buf := bytes.NewBufferString(e.Metric)
	if len(e.Params) > 0 {
		for _, param := range e.Params {
			buf.WriteRune(',')
			buf.WriteString(param)
		}
	}
	buf.WriteString(" value=")
	buf.WriteString(strconv.FormatInt(e.Value, 10))
	buf.WriteRune('\n')

	i.log.Increment("flush.size", int64(buf.Len()))

	before := time.Now()
	_, _ = i.writer.Write(buf.Bytes())
	i.log.Duration("flush.latency", time.Now().Sub(before))
}
