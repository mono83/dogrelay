package sentry

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const userAgent = "dogrelay/0.1"

// HTTPTransport is the default transport, delivering packets to Sentry via the
// HTTP API.
type HTTPTransport struct {
	*http.Client
}

// Send sends data to Sentry server
func (t *HTTPTransport) Send(url, authHeader string, packet Serializeable) error {
	if url == "" {
		return nil
	}

	body, contentType, err := packet.Serialize()
	if err != nil {
		return fmt.Errorf("error serializing packet: %v", err)
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("can't create new request: %v", err)
	}
	req.Header.Set("X-Sentry-Auth", authHeader)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", contentType)
	res, err := t.Do(req)
	if err != nil {
		return err
	}
	_, _ = io.Copy(ioutil.Discard, res.Body)
	_ = res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("raven: got http status %d", res.StatusCode)
	}
	return nil
}
