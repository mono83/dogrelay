package sentry

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

// SimplePacket is simplest packet, compatible with Sentry API
type SimplePacket struct {
	EventID     string            `json:"event_id"`
	Fingerprint []string          `json:"fingerprint,omitempty"`
	Message     string            `json:"message"`
	Timestamp   Timestamp         `json:"timestamp"`
	Logger      string            `json:"logger"`
	Route       string            `json:"transaction"`
	Level       Severity          `json:"level,omitempty"`
	Host        string            `json:"server_name,omitempty"`
	Release     string            `json:"release,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Extra       map[string]string `json:"extra,omitempty"`
	Exception   *SimpleExceptions `json:"exception,omitempty"`
}

// Serialize serializes packet
func (p SimplePacket) Serialize() (io.Reader, string, error) {
	packetJSON, err := json.Marshal(p)
	if err != nil {
		return nil, "", fmt.Errorf("error marshaling packet %+v to JSON: %v", p, err)
	}

	// Only deflate/base64 the packet if it is bigger than 1KB, as there is
	// overhead.
	if len(packetJSON) > 1000 {
		buf := &bytes.Buffer{}
		b64 := base64.NewEncoder(base64.StdEncoding, buf)
		deflate, _ := zlib.NewWriterLevel(b64, zlib.BestCompression)
		_, _ = deflate.Write(packetJSON)
		_ = deflate.Close()
		_ = b64.Close()
		return buf, "application/octet-stream", nil
	}
	return bytes.NewReader(packetJSON), "application/json", nil
}

// SimpleExceptions is simple exceptions list holder
type SimpleExceptions struct {
	Values []SimpleException `json:"values,omitempty"`
}

// SimpleException is simple exception data
type SimpleException struct {
	Message    string            `json:"value"`
	Class      string            `json:"type"`
	Stacktrace *SimpleStacktrace `json:"stacktrace,omitempty"`
}

// SimpleStacktrace is simple exception stacktrace
type SimpleStacktrace struct {
	Frames []SimpleFrame `json:"frames"`
}

// SimpleFrame is simple exception stack frame
type SimpleFrame struct {
	File     string `json:"filename,omitempty"`
	Function string `json:"function"`
	Line     int    `json:"lineno,omitempty"`
}
