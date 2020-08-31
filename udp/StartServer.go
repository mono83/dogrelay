package udp

import (
	"errors"
	"fmt"
	"github.com/mono83/dogrelay/metrics"
	"github.com/mono83/xray"
	"net"
	"sort"
	"strconv"
	"strings"
)

// StartServer starts plain UDP listener service
func StartServer(bind string, size int, clb func([]byte)) error {
	if size == 0 {
		size = 1024 * 8
	}

	if bind == "" {
		return errors.New("empty UDP address")
	}
	address, err := net.ResolveUDPAddr("udp", bind)
	if err != nil {
		return err
	}

	socket, err := net.ListenUDP("udp", address)
	if err != nil {
		return err
	}

	running := true
	log := xray.ROOT.Fork()
	// Listener
	go func() {
		for running {
			buf := make([]byte, size)
			rlen, _, err := socket.ReadFromUDP(buf)
			log.Inc("in.udp.count")
			log.Increment("in.udp.size", int64(rlen))
			if err != nil {
				// Connection error
				log.Inc("in.udp.error")
			} else {
				// Handling data
				go clb(buf[0:rlen])
			}
		}
	}()

	return nil
}

// StartMetricsServer starts UDP metrics listener service
func StartMetricsServer(bind string, size int, to func(metrics.Event)) error {
	return StartServer(
		bind,
		size,
		func(bts []byte) {
			// Parsing
			events, err := multiLineRead(bts)
			if err != nil {
				fmt.Println(err)
			}
			for _, e := range events {
				to(e)
			}
		},
	)
}

func multiLineRead(bts []byte) ([]metrics.Event, error) {
	str := string(bts)
	var result []metrics.Event
	for _, line := range strings.Split(str, "\n") {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			e, err := singleLineRead(line)
			if err != nil {
				return nil, err
			}

			result = append(result, e)
		}
	}

	return result, nil
}

func singleLineRead(str string) (metrics.Event, error) {
	chunks := strings.Split(strings.Replace(strings.Replace(str, ":", "|", 1), "#", "", 1), "|")
	if len(chunks) < 3 {
		return metrics.Event{}, errors.New("Invalid format string")
	}

	metric := chunks[0]
	value, err := strconv.ParseInt(chunks[1], 10, 64)
	if err != nil {
		return metrics.Event{}, err
	}
	typeString := chunks[2]
	var params []string
	if len(chunks) > 4 {
		// Making tags deduplication
		paramsMap := map[string]bool{}
		for _, v := range strings.Split(chunks[4], ",") {
			v = strings.TrimSpace(v)
			if len(v) == 0 {
				// Empty param
				continue
			}
			paramsMap[strings.Replace(v, ":", "=", 1)] = true
		}
		for p := range paramsMap {
			params = append(params, p)
		}
		sort.Strings(params)
	}

	switch typeString {
	case "c":
		return metrics.Event{EventType: metrics.TypeIncrement, Metric: metric, Value: value, Params: params}, nil
	case "g":
		return metrics.Event{EventType: metrics.TypeGauge, Metric: metric, Value: value, Params: params}, nil
	case "ms":
		return metrics.Event{EventType: metrics.TypeDuration, Metric: metric, Value: value, Params: params}, nil
	}

	return metrics.Event{}, fmt.Errorf("unsupported format %s", typeString)
}
