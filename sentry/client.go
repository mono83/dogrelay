package sentry

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Serializeable is interface of entities, that are able to be send to Sentry
type Serializeable interface {
	Serialize() (io.Reader, string, error)
}

// List of known errors
var (
	ErrMissingUser       = errors.New("raven: dsn missing public key and/or password")
	ErrMissingPrivateKey = errors.New("raven: dsn missing private key")
	ErrMissingProjectID  = errors.New("raven: dsn missing project id")
)

// NewClient builds and returns client for requested DSN
func NewClient(dsn string) (*Client, error) {
	cl := new(Client)
	if err := cl.SetDSN(dsn); err != nil {
		return nil, err
	}
	cl.queue = make(chan Serializeable)
	cl.transport = &HTTPTransport{Client: &http.Client{}}
	go cl.deliverLoop()

	return cl, nil
}

// Client is tiny Sentry client
type Client struct {
	url        string
	projectID  string
	authHeader string

	queue chan Serializeable

	transport *HTTPTransport
}

// SetDSN updates a client with a new DSN. It safe to call after and
// concurrently with calls to Report and Send.
func (client *Client) SetDSN(dsn string) error {
	if dsn == "" {
		return nil
	}

	uri, err := url.Parse(dsn)
	if err != nil {
		return err
	}

	if uri.User == nil {
		return ErrMissingUser
	}
	publicKey := uri.User.Username()
	secretKey, ok := uri.User.Password()
	if !ok {
		return ErrMissingPrivateKey
	}
	uri.User = nil

	if idx := strings.LastIndex(uri.Path, "/"); idx != -1 {
		client.projectID = uri.Path[idx+1:]
		uri.Path = uri.Path[:idx+1] + "api/" + client.projectID + "/store/"
	}
	if client.projectID == "" {
		return ErrMissingProjectID
	}

	client.url = uri.String()

	client.authHeader = fmt.Sprintf("Sentry sentry_version=4, sentry_key=%s, sentry_secret=%s", publicKey, secretKey)

	return nil
}

// Send places packet to outgoing queue
func (client *Client) Send(pkt Serializeable) {
	client.queue <- pkt
}

func (client *Client) deliverLoop() {
	for pkt := range client.queue {
		client.transport.Send(client.url, client.authHeader, pkt)
	}
}
