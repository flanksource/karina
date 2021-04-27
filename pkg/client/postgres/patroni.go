package postgres

import (
	"context"
	"fmt"
	"net/http"

	"github.com/flanksource/kommons/proxy"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PatroniCluster struct {
	Members []PatroniMember `json:"members"`
}

type PatroniMember struct {
	Name     string      `json:"name"`
	Host     string      `json:"host"`
	Port     int         `json:"port"`
	Role     string      `json:"role"`
	State    string      `json:"state"`
	URL      string      `json:"api_url"`
	Timeline int         `json:"timeline"`
	Lag      interface{} `json:"lag"`
}

func (db *PostgresDB) GetPatroniClient() (*http.Client, error) {
	client, _ := db.client.GetClientset()
	opts := metav1.ListOptions{LabelSelector: fmt.Sprintf("cluster-name=%s,spilo-role=master", db.Name)}
	pods, err := client.CoreV1().Pods(db.Namespace).List(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get master pod for cluster %s: %v", db.Name, err)
	}

	if len(pods.Items) != 1 {
		return nil, fmt.Errorf("expected 1 pod for spilo-role=master got %d", len(pods.Items))
	}

	dialer, err := db.client.GetProxyDialer(proxy.Proxy{
		Namespace:    db.Namespace,
		Kind:         "pods",
		ResourceName: pods.Items[0].Name,
		Port:         8008,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to get proxy dialer")
	}

	tr := &http.Transport{
		DialContext: dialer.DialContext,
	}

	httpClient := &http.Client{Transport: tr}

	return httpClient, nil
}

// func checkReplicaLag(p *platform.Platform, clusters ...string) error {
// 	for _, cluster := range clusters {
// 		patroniClient, err := GetPatroniClient(p, Namespace, cluster)
// 		if err != nil {
// 			return errors.Errorf("Failed to get patroni client to cluster %s", cluster)
// 		}
// 		response, err := patroniClient.Get("http://patroni/cluster")
// 		if err != nil {
// 			return errors.Errorf("Failed to get /cluster endpoint for cluster %s: %v", cluster, err)
// 		}
// 		defer response.Body.Close() // nolint: errcheck
// 		clusterResponse := &ClusterResponse{}
// 		err = json.NewDecoder(response.Body).Decode(&clusterResponse)
// 		if err != nil {
// 			return errors.Errorf("Failed to read response body for cluster %s: %v", cluster, err)
// 		}

// 		for _, m := range clusterResponse.Members {
// 			if m.State != "running" {
// 				return errors.Errorf("Expected state for cluster=%s node=%s to be 'running', got %s", cluster, m.Name, m.State)
// 			} else if m.Role == "replica" {
// 				iLag, ok := m.Lag.(int)
// 				if ok && iLag > 0 {
// 					return errors.Errorf("Expected replication lag for cluster=%s replica=%s to be 0, got %d", cluster, m.Name, m.Lag)
// 				} else if !ok {
// 					sLag, ok := m.Lag.(string)
// 					if ok && sLag != "" {
// 						return errors.Errorf("Expected replication lag for cluster=%s replica=%s to be 0, got %s", cluster, m.Name, m.Lag)
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }
