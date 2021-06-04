package elastic

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	elastic "github.com/olivere/elastic/v7"
)

type Query struct {
	Namespace  string
	Cluster    string
	Pod        string
	Count      int
	Query      string
	Since      string
	From, To   string
	Timestamps bool
}

type Fields struct {
	Cluster string `json:"cluster"`
}

type Kubernetes struct {
	Namespace string `json:"namespace"`
	Pod       Name   `json:"pod"`
	Container Name   `json:"container"`
}

type Message struct {
	Kubernetes Kubernetes `json:"kubernetes"`
	Host       Name       `json:"host"`
	Fields     Fields     `json:"fields"`
	Timestamp  string     `json:"@timestamp"`
	Message    string
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

	if query.From != "" {
		q.Must(elastic.NewRangeQuery("@timestamp").From(query.From).To(query.To))
	} else if query.Since != "" {
		q.Must(elastic.NewRangeQuery("@timestamp").From("now-" + query.Since).To("now"))
	}
	return q
}

func ExportLogs(p *platform.Platform, filebeatName string, query Query) error {
	var filebeat *types.Filebeat = nil

	for _, f := range p.Filebeat {
		if f.Name == filebeatName {
			filebeat = &f
		}
	}

	if filebeat == nil {
		return fmt.Errorf("failed to find filebeat with name %s", filebeatName)
	}

	p.Infof("Exporting logs from %s@%s", filebeat.Elasticsearch.User, filebeat.Elasticsearch.GetURL())
	var password string
	var err error
	_, password, err = p.GetEnvValue(filebeat.Elasticsearch.Password, "eck")
	if err != nil {
		_, password, err = p.GetEnvValue(filebeat.Elasticsearch.Password, metav1.NamespaceAll)
		if err != nil {
			return fmt.Errorf("unable to retrieve elasticsearch password for %s", filebeatName)
		}
	}
	es, err := elastic.NewSimpleClient(
		elastic.SetBasicAuth(filebeat.Elasticsearch.User, password),
		elastic.SetURL(filebeat.Elasticsearch.GetURL()),
	)
	if err != nil {
		return err
	}
	scroll := elastic.NewScrollService(es)

	pageSize := 5000
	if query.Count < pageSize {
		pageSize = query.Count
	}
	count := 0
	result, err := scroll.
		Index("filebeat-*").
		Size(pageSize).
		Sort("@timestamp", false).
		Query(query.ToQuery()).
		Do(context.Background())
	if err != nil {
		return err
	}

	for result.ScrollId != "" && count < query.Count {
		for _, hit := range result.Each(reflect.TypeOf(Message{})) {
			msg := hit.(Message)
			if query.Timestamps {
				fmt.Printf("[%s/%s/%s] %s %v\n", msg.Fields.Cluster, msg.Timestamp, msg.Kubernetes.Pod, msg.Kubernetes.Container, msg.Message)
			} else {
				fmt.Printf("[%s/%s/%s] %v\n", msg.Fields.Cluster, msg.Kubernetes.Pod, msg.Kubernetes.Container, msg.Message)
			}
			count++
			if count >= query.Count {
				break
			}
		}
		scollID := result.ScrollId
		result, err = scroll.ScrollId(scollID).Do(context.Background())
		if err != nil && errors.Is(err, io.EOF) {
			p.Infof("Exported %d results of %d total", count, result.TotalHits())
			return nil
		}

		if err != nil {
			time.Sleep(5 * time.Second)
			p.Infof("Retrying %s", err)
			result, err = scroll.ScrollId(scollID).Do(context.Background())
			if err != nil && errors.Is(err, io.EOF) {
				p.Infof("Exported %d results of %d total", count, result.TotalHits())
				return nil
			}
		}
		if err != nil {
			return err
		}

		p.Infof("Exported %d results of %d total", count, result.TotalHits())
	}

	return nil
}
