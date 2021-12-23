package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/mono83/xray"
	"github.com/mono83/xray/args"
	"strings"
	"time"
)

// Client is a wrapper over Elasticsearch client
type Client struct {
	client      *elasticsearch.Client
	logger      xray.Ray
	indexFormat string
}

// NewClient constructs new Elasticsearch client
func NewClient(addresses []string, indexFormat, user, pass string) (*Client, error) {
	for _, addr := range addresses {
		xray.BOOT.Info("ElasticSearch address is :addr", args.Addr(addr))
	}
	cfg := elasticsearch.Config{
		Addresses:            addresses,
		Username:             user,
		Password:             pass,
		EnableRetryOnTimeout: true,
		//DiscoverNodesOnStart: true,
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	res, err := es.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	infoResult := map[string]interface{}{}
	if err := json.NewDecoder(res.Body).Decode(&infoResult); err != nil {
		return nil, err
	}
	xray.BOOT.Info("Elasticsearch client: :version", args.String{N: "version", V: elasticsearch.Version})
	xray.BOOT.Info("Elasticsearch server: :version", args.String{N: "version", V: infoResult["version"].(map[string]interface{})["number"].(string)})

	return &Client{
		client:      es,
		logger:      xray.ROOT.Fork().WithLogger("elastic").WithMetricPrefix("elastic"),
		indexFormat: indexFormat,
	}, nil
}

// CreateIndexTemplate creates standard ES template for logs.
func (c *Client) CreateIndexTemplate() error {
	req := esapi.IndicesPutTemplateRequest{
		Name:   "logstash",
		Create: refBool(true),
		Body:   strings.NewReader(indexTemplateString),
	}
	res, err := req.Do(context.TODO(), c.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e esErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&e); err == nil {
			if strings.HasSuffix(e.Error.Reason, "already exists") {
				xray.BOOT.Info("Index template already exists")
				return nil
			} else {
				xray.BOOT.Error("Reason - :err", args.Error{Err: e.Error})
			}
		}
		xray.BOOT.Error("Template creation failed - :err", args.Error{Err: errors.New(res.Status())})
	} else {
		xray.BOOT.Info("Template created")
	}
	return nil
}

// Write writes data to ES
func (c *Client) Write(b []byte) error {
	req := esapi.IndexRequest{
		Index:      time.Now().Format(c.indexFormat),
		Body:       bytes.NewReader(b),
		Refresh:    "true",
		DocumentID: "",
	}

	// Performing request
	res, err := req.Do(context.Background(), c.client)
	if err != nil {
		c.logger.Error("WTF - :err", args.Error{Err: err})
		c.logger.Inc("error", args.Type("io"))
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		c.logger.Error(res.Status())
		c.logger.Inc("error", args.Type("derived"))
		return fmt.Errorf("error indexing " + res.Status())
	}

	fmt.Println(time.Now().Format(c.indexFormat))
	fmt.Println(string(b))

	return nil
}

func refBool(b bool) *bool {
	return &b
}
