package elasticsearch

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/api/elasticsearch"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	"github.com/flanksource/kommons/proxy"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Elasticsearch == nil || p.Elasticsearch.Disabled {
		return
	}

	client, err := p.GetClientset()
	if err != nil {
		test.Failf("Elasticsearch", "Failed to get k8s client: %v", err)
		return
	}
	kommons.TestNamespace(client, Namespace, test)

	clusterName := "logs"
	userName := "elastic"

	pod, err := p.GetFirstPodByLabelSelector(Namespace, fmt.Sprintf("common.k8s.elastic.co/type=elasticsearch,elasticsearch.k8s.elastic.co/cluster-name=%s", clusterName))
	if err != nil {
		test.Failf("Elasticsearch", "Unable to find elastic pod")
		return
	}

	dialer, _ := p.GetProxyDialer(proxy.Proxy{
		Namespace:    Namespace,
		Kind:         "pods",
		ResourceName: pod.Name,
		Port:         9200,
	})
	tr := &http.Transport{
		DialContext:     dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}

	secret := p.GetSecret(Namespace, fmt.Sprintf("%s-es-%s-user", clusterName, userName))
	if secret == nil {
		test.Failf("Elasticsearch", "Unable to get password for %s user %v", userName, err)
		return
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://%s-es-http/_cluster/health", clusterName), nil)
	req.Header.Add("Authorization", "Basic "+basicAuth(userName, string((*secret)[userName])))

	resp, err := httpClient.Do(req)
	if err != nil {
		test.Failf("Elasticsearch", "Failed to get cluster health: %v", err)
		return
	}
	health := elasticsearch.Health{}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &health); err != nil {
		test.Failf("Elasticsearch", "Failed to unmarshall :%v", err)
	} else if health.Status == elasticsearch.GreenHealth {
		test.Passf("Elasticsearch", "elasticsearch cluster is: %s", health)
	} else {
		test.Failf("Elasticsearch", "elasticsearch cluster is: %s", health)
	}
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
