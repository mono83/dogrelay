package udp

import (
	"errors"
	"fmt"
	"github.com/mono83/dogrelay/metrics"
	"net"
	"sort"
	"strconv"
	"strings"
)

// StartServer starts UDP listener service
func StartServer(bind string, size int, to func(metrics.Event)) error {
	if size == 0 {
		size = 1024 * 8
	}

	if bind == "" {
		return errors.New("Empty UDP address")
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

	// Listener
	go func() {
		for running {
			buf := make([]byte, size)
			rlen, _, err := socket.ReadFromUDP(buf)
			if err != nil {
				// Connection error
			} else {
				go func(bts []byte) {
					// Parsing
					events, err := multiLineRead(bts)
					if err != nil {
						fmt.Println(err)
					}
					for _, e := range events {
						to(e)
					}
				}(buf[0:rlen])
			}
		}
	}()

	return nil
}

func multiLineRead(bts []byte) ([]metrics.Event, error) {
	str := string(bts)
	result := []metrics.Event{}
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
	params := []string{}
	if len(chunks) > 4 {
		params = strings.Split(chunks[4], ",")
		for k, v := range params {
			params[k] = strings.Replace(v, ":", "=", 1)
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

	return metrics.Event{}, fmt.Errorf("Unsupported format %s", typeString)
}