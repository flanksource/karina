package elasticsearch

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/api/elasticsearch"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/k8s/proxy"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Elasticsearch == nil || p.Elasticsearch.Disabled {
		test.Skipf("Elasticsearch", "elastichsearch is not installed or enabled")
		return
	}

	client, err := p.GetClientset()
	if err != nil {
		test.Failf("Elasticsearch", "Failed to get k8s client: %v", err)
		return
	}
	k8s.TestNamespace(client, Namespace, test)

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

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s-es-http/_cluster/health", clusterName), nil)
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

func redirectPolicyFunc(req *http.Request, via []*http.Request) error {

	return nil
}
