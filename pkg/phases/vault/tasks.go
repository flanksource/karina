package vault

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/flanksource/commons/certs"
	"github.com/hashicorp/vault/api"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Init(p *platform.Platform) error {
	if p.Vault == nil || p.Vault.Disabled {
		p.Infof("Vault is not configured or disabled. Nothing to be done")
		return nil
	}
	p.Infof("Waiting for vault/vault-0 to be running")
	if err := p.WaitForPod("vault", "vault-0", 300*time.Second, v1.PodRunning); err != nil {
		return err
	}
	p.Infof("Waiting for vault service to be listening")
	if err := p.WaitForPodCommand("vault", "vault-0", "vault", 120*time.Second, "/bin/sh", "-c", "netstat -tln | grep 8200"); err != nil {
		return err
	}

	stdout, stderr, err := p.ExecutePodf("vault", "vault-0", "vault", "/bin/vault", "operator", "init", "-tls-skip-verify", "-status")

	p.Debugf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
	if strings.Contains(stdout, "Vault is initialized") {
		p.Infof("Vault is already initialized, configuring")
	}
	if p.Vault.Token == "" {
		p.Infof("Vault is not initialized, initializing")
		stdout, stderr, err = p.ExecutePodf("vault", "vault-0", "vault", "/bin/vault", "operator", "init", "-tls-skip-verify")
		p.Infof("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)

		tokens := regexp.MustCompile(`(?m)Initial Root Token: (s.\w+).`).FindStringSubmatch(stdout)
		if tokens == nil {
			return fmt.Errorf("no root token found")
		}

		p.Vault.Token = tokens[0]
	}
	config := &api.Config{
		Address: "https://vault." + p.Domain + ":443",
	}
	_ = config.ConfigureTLS(&api.TLSConfig{
		Insecure: true,
	})

	client, err := api.NewClient(config)
	if err != nil {
		return err
	}

	client.SetToken(p.Vault.Token)
	if err := configurePKI(client, p); err != nil {
		return err
	}
	if err := configureLdap(client, p); err != nil {
		return err
	}

	yes := true
	secret, err := client.Auth().Token().Create(&api.TokenCreateRequest{
		Policies:    []string{"signer"},
		DisplayName: "karina",
		Metadata: map[string]string{
			"value": "key",
		},
		Lease:     "8760h",
		Renewable: &yes,
		TTL:       "8760h",
		Period:    "8760h", //1y
	})

	if err != nil {
		return err
	}
	p.Infof("Token: %s", secret.Auth.ClientToken)
	for name, policy := range p.Vault.Policies {
		if _, err := client.Logical().Write("sys/policy/"+name, map[string]interface{}{
			"policy": policy.String(),
		}); err != nil {
			return err
		}
	}

	// ExtraConfig is an escape hatch that allows writing to arbritrary vault paths
	for path, config := range p.Vault.ExtraConfig {
		if _, err := client.Logical().Write(path, config); err != nil {
			return fmt.Errorf("error writing to %s: %v", path, err)
		}
	}
	return nil
}

func configurePKI(client *api.Client, p *platform.Platform) error {
	mounts, err := client.Sys().ListMounts()
	if err != nil {
		return err
	}

	if _, ok := mounts["pki/"]; !ok {
		if err := client.Sys().Mount("pki", &api.MountInput{
			Type: "pki",
			Config: api.MountConfigInput{
				MaxLeaseTTL: "43800h", // 5 years
			},
		}); err != nil {
			return err
		}
	}

	ingress := p.GetIngressCA()
	switch ingress := ingress.(type) {
	case *certs.Certificate:
		if _, err := client.Logical().Write("pki/config/ca", map[string]interface{}{
			"pem_bundle": string(ingress.EncodedCertificate()) + "\n" + string(ingress.EncodedPrivateKey()),
		}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown CA type %v", ingress)
	}

	for role, config := range p.Vault.Roles {
		if _, err := client.Logical().Write("pki/roles/"+role, config); err != nil {
			return err
		}
	}

	return nil
}

func configureLdap(client *api.Client, p *platform.Platform) error {
	auths, err := client.Sys().ListAuth()
	if err != nil {
		return err
	}

	if _, exists := auths["ldap/"]; !exists {
		if err := client.Sys().EnableAuthWithOptions("ldap", &api.EnableAuthOptions{
			Type: "ldap",
		}); err != nil {
			return err
		}
	}

	if _, err := client.Logical().Write("auth/ldap/config", map[string]interface{}{
		"url":          p.Ldap.GetConnectionURL(),
		"binddn":       p.Ldap.Username,
		"bindpass":     p.Ldap.Password,
		"userdn":       p.Ldap.UserDN,
		"groupdn":      p.Ldap.GroupDN,
		"groupfilter":  fmt.Sprintf("(&(objectClass=%s)(member={{.UserDN}}))", p.Ldap.GroupObjectClass),
		"groupattr":    p.Ldap.GroupNameAttr,
		"userattr":     "cn",
		"insecure_tls": "true",
		"starttls":     "true",
	}); err != nil {
		return err
	}

	for group, policies := range p.Vault.GroupMappings {
		if _, err := client.Logical().Write("auth/ldap/groups/"+group, map[string]interface{}{
			"policies": strings.Join(policies, ","),
		}); err != nil {
			return err
		}
	}

	return nil
}
