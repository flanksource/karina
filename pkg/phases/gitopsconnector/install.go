package gitopsconnector

import (
	"fmt"
	"strings"

	"github.com/flanksource/karina/pkg/platform"
	v1 "k8s.io/api/core/v1"
)

const (
	Spec      = "gitops-connector.yaml"
	Namespace = "gitops-connector"
)

func Install(p *platform.Platform) error {
	if p.GitOpsConnector.IsDisabled() {
		if err := p.DeleteSpecs(v1.NamespaceAll, Spec); err != nil {
			p.Errorf("failed to delete specs: %v", err)
		}
		return nil
	}

	if err := p.CreateOrUpdateNamespace(Namespace, nil, nil); err != nil {
		return fmt.Errorf("failed to create/update namespace: %v", err)
	}

	// The gitops-connector expects these to be upper case, handle configuration when it is not
	p.GitOpsConnector.GitOpsOperatorType = strings.ToUpper(p.GitOpsConnector.GitOpsOperatorType)
	p.GitOpsConnector.GitRepositoryType = strings.ToUpper(p.GitOpsConnector.GitRepositoryType)
	p.GitOpsConnector.CICDOrchestratorType = strings.ToUpper(p.GitOpsConnector.CICDOrchestratorType)

	config := map[string]string{
		"GITOPS_OPERATOR_TYPE":   p.GitOpsConnector.GitOpsOperatorType,
		"GIT_REPOSITORY_TYPE":    p.GitOpsConnector.GitRepositoryType,
		"CICD_ORCHESTRATOR_TYPE": p.GitOpsConnector.CICDOrchestratorType,
		"GITOPS_APP_URL":         p.GitOpsConnector.GitOpsAppURL,
	}

	if p.GitOpsConnector.GitRepositoryType == "AZDO" {
		config["AZDO_GITOPS_REPO_NAME"] = p.GitOpsConnector.GitRepository.ManifestsRepo
		config["AZDO_ORG_URL"] = p.GitOpsConnector.GitRepository.OrgURL
		if p.GitOpsConnector.GitRepository.PullRequestRepo != "" {
			config["AZDO_PR_REPO_NAME"] = p.GitOpsConnector.GitRepository.PullRequestRepo
		}
	} else {
		config["GITHUB_GITOPS_MANIFEST_REPO_NAME"] = p.GitOpsConnector.GitRepository.ManifestsRepo
		config["GITHUB_ORG_URL"] = p.GitOpsConnector.GitRepository.OrgURL
		config["GITHUB_GITOPS_REPO_NAME"] = p.GitOpsConnector.GitRepository.PullRequestRepo
	}

	if err := p.CreateOrUpdateConfigMap("gitops-connector-cm", Namespace, config); err != nil {
		return err
	}

	if err := p.CreateOrUpdateConfigMap("gitops-connector-subscribers-config", Namespace, p.GitOpsConnector.Subscribers); err != nil {
		return err
	}

	if err := p.ApplySpecs(Namespace, Spec); err != nil {
		return err
	}
	return nil
}
