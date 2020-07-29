package externalclusters

import (
	"fmt"
	"net/url"

	"github.com/flanksource/commons/certs"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/karina/pkg/k8s"
)

// ExternalClusters is a map of clusterName: clusterApiEndpoints
// with convenience methods.
type ExternalClusters map[string]string

// AddSelf adds the default internal k8s API endpoint under the given cluster name
// to describe "internal" access.
func (ec *ExternalClusters) AddSelf(name string) {
	(*ec)[name] = "https://kubernetes.default"
}

// clientFunc is an internal convenience type used for k8s.Client function references
type clientFunc func(c *k8s.Client) error

// ApplySpecs applies the given specs to each external cluster.
// Errors result in log output to the given logger.
// An ExternalClusters containing only successfully processed
// external clusters is returned at completion.
func (ec *ExternalClusters) ApplySpecs(ca *certs.Certificate, logger logger.Logger, specs ...string) (*ExternalClusters, error) {
	if len(*ec) < 1 {
		return nil, fmt.Errorf("no external clusters configured")
	}
	applyFunc := func(c *k8s.Client) error {
		return c.ApplyText("default", specs...)
	}
	return ec.processClusters(ca, logger, applyFunc)
}

// DeleteSpecs seletes the given specs on each external cluster.
// Errors result in log output to the given logger.
// An ExternalClusters containing only successfully processed
// external clusters is returned at completion.
func (ec *ExternalClusters) DeleteSpecs(ca *certs.Certificate, logger logger.Logger, specs ...string) (*ExternalClusters, error) {
	if len(*ec) < 1 {
		return nil, fmt.Errorf("no external clusters configured")
	}
	applyFunc := func(c *k8s.Client) error {
		return c.DeleteText("default", specs...)
	}
	return ec.processClusters(ca, logger, applyFunc)
}

// processClusters invokes the given clientFunc on each external cluster using
// a k8s.Client constructed for each given endpoint.
// Errors result in log output to the given logger.
// An ExternalClusters containing only successfully processed
// external clusters is returned at completion.
func (ec *ExternalClusters) processClusters(ca *certs.Certificate, logger logger.Logger, cf clientFunc) (*ExternalClusters, error) {
	if len(*ec) < 1 {
		return nil, fmt.Errorf("no external clusters configured")
	}
	clusters := ExternalClusters{}
	for name, apiEndpoint := range *ec {
		logger.Infof("processing external cluster %v with endpoint: %v", name, apiEndpoint)
		u, err := url.Parse(apiEndpoint)
		if err != nil {
			logger.Errorf("Unable to parse external cluster endpoint URL: %v", apiEndpoint)
			continue
			// failing to process this external cluster - try the next one
		}
		if u.Port() != "6443" {
			logger.Errorf("Only port 6443 supported for external cluster endpoint URLs: %v", apiEndpoint)
			// because k8s.GetExternalClient
			continue
			// failing to process this external cluster - try the next one
		}
		logger.Debugf("External endpoint host: %v", u.Hostname())

		client := k8s.GetExternalClient(logger, name, u.Hostname(), ca)

		err = cf(client)
		if err != nil {
			logger.Errorf("Error processing cluster: %v", err)
			continue
			// failing to add this external cluster - try the next one
		}
		// if the cluster was configured we return it in the result map
		clusters[name] = apiEndpoint
	}
	return &clusters, nil
}
