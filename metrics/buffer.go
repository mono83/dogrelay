package metrics

import (
	"sort"
	"sync"
)

// Buffer structure contains buffered information about
// metrics events and can be used to provide aggregated one
type Buffer struct {
	lock sync.Mutex

	percentiles []int
	compatMode  bool

	prototypes map[string]Event
	counters   map[string]int64
	gauges     map[string]int64
	durations  map[string][]int64
}

// NewBuffer builds new Buffer
func NewBuffer(percentiles []int, compatMode bool) *Buffer {
	return &Buffer{
		percentiles: percentiles,
		compatMode:  compatMode,
		prototypes:  map[string]Event{},
		counters:    map[string]int64{},
		gauges:      map[string]int64{},
		durations:   map[string][]int64{},
	}
}

// Add registers new event
func (b *Buffer) Add(e Event) {
	// Reading key
	key := e.Key()

	b.lock.Lock()
	defer b.lock.Unlock()

	if _, ok := b.prototypes[key]; !ok {
		// Storing prototype event
		b.prototypes[key] = e
	}

	switch e.EventType {
	case TypeIncrement:
		prev, ok := b.counters[key]
		if ok {
			b.counters[key] = prev + e.Value
		} else {
			b.counters[key] = e.Value
		}
	case TypeGauge:
		b.gauges[key] = e.Value
	case TypeDuration:
		b.durations[key] = append(b.durations[key], e.Value)
	}
}

// Flush flushes buffered events into aggregated list
func (b *Buffer) Flush() []Event {
	result := []Event{}

	// First lock - working with counters and gauges, and copying
	// durations to local variable
	b.lock.Lock()
	for k, v := range b.gauges {
		if b.compatMode {
			result = append(result, b.prototypes[k].WithValueSuffix(v, ".gauge"))
		} else {
			result = append(result, b.prototypes[k].WithValue(v))
		}
	}
	for k, v := range b.counters {
		if b.compatMode {
			result = append(result, b.prototypes[k].WithValueSuffix(v, ".counter"))
		} else {
			result = append(result, b.prototypes[k].WithValue(v))
		}
	}

	b.counters = map[string]int64{}
	local := b.durations
	b.durations = map[string][]int64{}
	b.lock.Unlock()

	if len(local) == 0 {
		return result
	}

	// Reading all prototypes, used by durations
	prototypes := make(map[string]Event, len(local))
	b.lock.Lock()
	for k := range local {
		prototypes[k] = b.prototypes[k]
	}
	b.lock.Unlock()

	// Flattening
	for k, v := range local {
		result = append(result, b.flatten(prototypes[k], int64arr(v))...)
	}

	return result
}

func (b *Buffer) flatten(proto Event, values int64arr) []Event {
	result := []Event{}
	sort.Sort(values)

	compatPrefix := ""
	if b.compatMode {
		compatPrefix = ".timer"
	}

	// Calculating sum and avg
	var sum, avg int64
	count := len(values)
	for _, v := range values {
		sum += v
	}
	avg = sum / int64(count)

	result = append(result, proto.WithValueSuffix(int64(count), compatPrefix+".count"))
	result = append(result, proto.WithValueSuffix(sum, compatPrefix+".sum"))
	result = append(result, proto.WithValueSuffix(avg, compatPrefix+".avg"))
	result = append(result, proto.WithValueSuffix(values[0], compatPrefix+".min"))
	result = append(result, proto.WithValueSuffix(values[len(values)-1], compatPrefix+".max"))

	// Calculating percentiles
	var percSum, upperSum int64
	for _, perc := range b.percentiles {
		percIndex := int(perc * count / 100)
		meanList := values[0 : percIndex+1]
		upperList := values[percIndex:]
		percSum = 0
		upperSum = 0
		if len(meanList) > 0 {
			for _, v := range meanList {
				percSum += v
			}
			result = append(result, proto.WithValueSuffixI(percSum/int64(len(meanList)), compatPrefix+".mean_", perc))
		}
		if len(upperList) > 0 {
			for _, v := range upperList {
				upperSum += v
			}
			result = append(result, proto.WithValueSuffixI(upperSum/int64(len(upperList)), compatPrefix+".upper_", perc))
		} else {
			result = append(result, proto.WithValueSuffixI(values[len(values)-1], compatPrefix+".upper_", perc))
		}
		result = append(result, proto.WithValueSuffixI(meanList[len(meanList)-1], compatPrefix+".perc_", perc))
	}

	return result
}

type int64arr []int64

func (a int64arr) Len() int {
	return len(a)
}
func (a int64arr) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a int64arr) Less(i, j int) bool {
	return a[i] < a[j]
}
