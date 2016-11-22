package udp

import (
	"github.com/mono83/dogrelay/metrics"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMultiLineRead(t *testing.T) {
	assert := assert.New(t)

	events, err := multiLineRead([]byte("foo:1|c\nbar:2|g\n"))
	if assert.NoError(err) {
		assert.Len(events, 2)
	}
}

func TestSingleLineRead(t *testing.T) {
	assert := assert.New(t)

	event, err := singleLineRead("foo:1|c")
	if assert.NoError(err) {
		assert.Equal("foo", event.Metric)
		assert.Equal(int64(1), event.Value)
		assert.Equal(metrics.TypeIncrement, event.EventType)
		assert.Len(event.Params, 0)
	}
	event, err = singleLineRead("bar:-7|g|@1.0")
	if assert.NoError(err) {
		assert.Equal("bar", event.Metric)
		assert.Equal(int64(-7), event.Value)
		assert.Equal(metrics.TypeGauge, event.EventType)
		assert.Len(event.Params, 0)
	}
	event, err = singleLineRead("latency:344|ms")
	if assert.NoError(err) {
		assert.Equal("latency", event.Metric)
		assert.Equal(int64(344), event.Value)
		assert.Equal(metrics.TypeDuration, event.EventType)
		assert.Len(event.Params, 0)
	}
	event, err = singleLineRead("users.online:800|g|@0.5|#country:china,server:china-1")
	if assert.NoError(err) {
		assert.Equal("users.online", event.Metric)
		assert.Equal(int64(800), event.Value)
		assert.Equal(metrics.TypeGauge, event.EventType)
		if assert.Len(event.Params, 2) {
			assert.Equal("country=china", event.Params[0])
			assert.Equal("server=china-1", event.Params[1])
			assert.Equal("g\tusers.online\tcountry=china\tserver=china-1", event.Key())
		}
	}
}
