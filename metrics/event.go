package metrics

import (
	"fmt"
	"strings"
)

// Type constants
const (
	TypeIncrement byte = 'i'
	TypeGauge     byte = 'g'
	TypeDuration  byte = 'd'
)

// Event used by buffer
type Event struct {
	EventType byte
	Value     int64
	Metric    string
	Params    []string
}

// Key method returns key for hash map
func (e Event) Key() string {
	return string(e.EventType) + "\t" + e.Metric + "\t" + strings.Join(e.Params, "\t")
}

// WithValueSuffix returns new Event with new value and suffix
func (e Event) WithValueSuffix(value int64, suffix string) Event {
	return Event{
		EventType: e.EventType,
		Value:     value,
		Metric:    e.Metric + suffix,
		Params:    e.Params,
	}
}

// WithValueSuffixI returns new Event with new value and suffix
func (e Event) WithValueSuffixI(value int64, suffix string, percentile int) Event {
	return Event{
		EventType: e.EventType,
		Value:     value,
		Metric:    fmt.Sprintf("%s%s%d", e.Metric, suffix, percentile),
		Params:    e.Params,
	}
}

// WithValue returns new Event with changed value
func (e Event) WithValue(value int64) Event {
	return Event{
		EventType: e.EventType,
		Value:     value,
		Metric:    e.Metric,
		Params:    e.Params,
	}
}
