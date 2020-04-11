package elastic

import (
	"context"
	"fmt"
	"reflect"

	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/olivere/elastic/v7"
	log "github.com/sirupsen/logrus"
)

type Query struct {
	Namespace string
	Cluster   string
	Pod       string
	Count     int
	Query     string
}

type Message struct {
	Namespace string `json:"namespace"`
	Pod       Name   `json:"pod"`
	Host      Name   `json:"host"`
	Timestamp string `json:"@timestamp"`
	Message   string
}
type Name struct {
	Name string
}

func (n Name) String() string {
	return n.Name
}

func (query Query) ToQuery() elastic.Query {
	q := elastic.NewBoolQuery()
	if query.Namespace != "" {
		q.Must(elastic.NewMatchPhraseQuery("kubernetes.namespace", query.Namespace))
	}
	if query.Pod != "" {
		q.Must(elastic.NewMatchPhraseQuery("kubernetes.pod.name", query.Pod))
	}
	if query.Cluster != "" {
		q.Must(elastic.NewMatchPhraseQuery("fields.cluster", query.Cluster))
	}
	if query.Query != "" {
		q.Must(elastic.NewQueryStringQuery(query.Query))
	}
	return q
}

func ExportLogs(p *platform.Platform, query Query, dst string) error {
	es, err := elastic.NewSimpleClient(
		elastic.SetBasicAuth(p.Filebeat.Elasticsearch.User, p.Filebeat.Elasticsearch.Password),
		elastic.SetURL(p.Filebeat.Elasticsearch.GetURL()),
	)
	if err != nil {
		return err
	}

	result, err := es.Search().
		Index("filebeat-*").
		Size(query.Count).
		Sort("@timestamp", false).
		Query(query.ToQuery()).
		Do(context.Background())
	if err != nil {
		return err
	}

	for _, hit := range result.Each(reflect.TypeOf(Message{})) {
		msg := hit.(Message)
		fmt.Printf("[%s/%s/%s] %v\n", console.Greenf(query.Pod), msg.Namespace, msg.Pod, msg.Message)
	}
	log.Infof("Export %d results of %d total", len(result.Hits.Hits), result.TotalHits())
	return nil
}
