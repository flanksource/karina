package vault

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"

	"github.com/flanksource/commons/certs"
	"github.com/hashicorp/vault/api"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Init(p *platform.Platform) error {
	if err := p.WaitForPod("vault", "vault-0", 120*time.Second, v1.PodRunning); err != nil {
		return err
	}

	if err := p.WaitForPodCommand("vault", "vault-0", "vault", 20*time.Second, "/bin/sh", "-c", "netstat -tln | grep 8200"); err != nil {
		return err
	}

	stdout, stderr, err := p.ExecutePodf("vault", "vault-0", "vault", "/bin/vault", "operator", "init", "-tls-skip-verify", "-status")

	log.Debugf("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)
	if strings.Contains(stdout, "Vault is initialized") {
		log.Infof("Vault is already initialized, configuring")
		// return nil
	}
	if p.Vault.Token == "" {
		log.Infof("Vault is not initialized, initializing")
		stdout, stderr, err = p.ExecutePodf("vault", "vault-0", "vault", "/bin/vault", "operator", "init", "-tls-skip-verify")
		log.Infof("stdout: %s, stderr: %s, err: %v", stdout, stderr, err)

		tokens := regexp.MustCompile(`(?m)Initial Root Token: (\w+).`).FindStringSubmatch(stdout)
		if tokens == nil {
			return fmt.Errorf("Not root token found")
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
		mounts, _ = client.Sys().ListMounts()
	}

	fmt.Printf("%v", *mounts["pki/"])

	ingress := p.GetIngressCA()
	switch ingress := ingress.(type) {
	case *certs.Certificate:
		if _, err := client.Logical().Write("pki/config/ca", map[string]interface{}{
			"pem_bundle": string(ingress.EncodedCertificate()) + "\n" + string(ingress.EncodedPrivateKey()),
		}); err != nil {
			return err
		}

	default:
		fmt.Printf("Unknown CA type %v", ingress)
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
		"groupfilter":  fmt.Sprintf("(&(objectClass=%s)({{.UserDN}}))", p.Ldap.GroupObjectClass),
		"groupattr":    p.Ldap.GroupNameAttr,
		"userattr":     "cn",
		"insecure_tls": "true",
		"starttls":     "true",
	}); err != nil {
		return err
	}
	return nil
}
