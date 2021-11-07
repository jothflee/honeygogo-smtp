package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/jothflee/honeygogo/core"
)

func NewESBackend(index string) Backend {
	esBackend := ESBackend{
		Index: index,
	}
	esBackend.Connect()
	return &esBackend
}

type ESBackend struct {
	Backend
	Index  string
	client *elasticsearch.Client
	bulk   esutil.BulkIndexer
}

func (esb *ESBackend) Connect() {
	esAddr := os.Getenv("ELASTICSEARCH_URL")
	if esAddr != "" {
		index := esb.Index
		log.Debugf("ELASTICSEARCH_URL env configured to: %s", esAddr)
		// TODO: make more complex
		es, err := elasticsearch.NewDefaultClient()
		if err != nil {
			log.Trace("Error creating the es client: %s", err)
		} else {
			req := esapi.InfoRequest{}
			resp, err := req.Do(context.Background(), es)
			if err != nil {
				log.Errorf(err.Error())
			} else {
				log.Info(resp.String())
				// create honeygogo index
				createESIndex(es, index)
				// create bulk indexer
				indexer, _ := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
					Client:        es,
					Index:         index,
					FlushInterval: 1 * time.Second,
				})

				esb.client = es
				esb.bulk = indexer
			}

		}
	} else {
		log.Trace("ELASTICSEARCH_URL env not configured.")
	}

}
func (esb *ESBackend) OnMessage(msg core.MessageMeta) {

	if esb.client == nil {
		esb.Connect()
	}

	if esb.client != nil {
		b, err := json.Marshal(msg)
		if err == nil {
			esb.bulk.Add(
				context.Background(),
				esutil.BulkIndexerItem{
					Action: "index",
					Body:   bytes.NewReader(b),
				},
			)
		} else {
			log.Error(err)
		}
	}

}
func (esb *ESBackend) Close() {

	if esb.bulk != nil {
		esb.bulk.Close(context.Background())
	}
}

func createESIndex(client *elasticsearch.Client, index string) {
	req := esapi.IndicesCreateRequest{
		Index:  index,
		Human:  true,
		Pretty: true,
		Body: strings.NewReader(`{
		"settings":{},
		"mappings":{
			"properties": {
				"loc": {
					"type": "geo_point"
				},
				"ts": {
					"type": "date",
					"format": "epoch_millis"
				},
				"to_addr":{
					"type": "ip"
				},
				"from_addr":{
					"type": "ip"
				}
			}
		}
	}`),
	}
	resp, err := req.Do(context.Background(), client)
	if err != nil {
		log.Errorf(err.Error())
	}
	if resp.StatusCode >= 300 {
		log.Errorf(resp.String())
	}

}
