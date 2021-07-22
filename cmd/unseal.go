package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/karina/pkg/ca"
	"github.com/flanksource/karina/pkg/platform"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	sealedSecretsNamespace = "sealed-secrets"
	privateKeyFile         = ".kubeseal.sealed-secrets-key.json"
)

func getOfflineCertificate(p *platform.Platform) (corev1.Secret, error) {
	ca, err := ca.ReadCA(p.SealedSecrets.Certificate)
	if err != nil {
		return corev1.Secret{}, err
	}
	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sealed-secrets-key",
			Namespace: sealedSecretsNamespace,
		},
		Data: ca.AsTLSSecret(),
	}
	return secret, nil
}

func getOnlineCertificate(p *platform.Platform) (corev1.Secret, error) {
	client, err := p.GetClientset()
	if err != nil {
		return corev1.Secret{}, err
	}
	secretsInterface := client.CoreV1().Secrets(sealedSecretsNamespace)
	secrets, err := secretsInterface.List(context.TODO(), metav1.ListOptions{
		LabelSelector: "sealedsecrets.bitnami.com/sealed-secrets-key=active",
	})
	if err != nil {
		return corev1.Secret{}, err
	}
	if len(secrets.Items) < 1 {
		return corev1.Secret{}, errors.New("no sealed-secrets active private key set")
	}
	secret := secrets.Items[0]
	// Required because k8s API client does not set it automatically.
	typeMeta := metav1.TypeMeta{
		Kind:       "Secret",
		APIVersion: "v1",
	}
	secret.TypeMeta = typeMeta
	return secret, nil
}

var Unseal = &cobra.Command{
	Use:   "unseal",
	Short: "Unseal a secret using sealed-secrets",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		p := getPlatform(cmd)
		filePath, err := filepath.Abs(privateKeyFile)
		if err != nil {
			log.Fatalf("Unable to resolve file path for %s: %v", privateKeyFile, err)
		}
		flags := []string{
			"--recovery-unseal",
			getFlagString("format", cmd),
			fmt.Sprintf("--recovery-private-key %s", filePath),
		}
		if !p.SealedSecrets.Disabled && p.SealedSecrets.Certificate != nil {
			if p.SealedSecrets.Certificate.Cert == "" {
				log.Fatalf("Sealed-secrets certificate not provided in config")
			}
			flags = append(flags, "--cert", p.SealedSecrets.Certificate.Cert)
		}
		isOffline, _ := cmd.Flags().GetBool("offline")
		var secret corev1.Secret
		if isOffline {
			if p.SealedSecrets.Certificate.PrivateKey == "" {
				log.Fatalf("Sealed-secrets private key not provided in config")
			}
			secret, err = getOfflineCertificate(p)
			if err != nil {
				log.Fatalf("Unable to get sealed-secrets private key from file: %v", err)
			}
		} else {
			secret, err = getOnlineCertificate(p)
			if err != nil {
				log.Fatalf("Unable to get sealed-secrets active private key from cluster: %v", err)
			}
		}
		file, err := json.Marshal(secret)
		if err != nil {
			log.Fatalf("Unable to get marshal secret to JSON: %v", err)
		}
		err = ioutil.WriteFile(privateKeyFile, file, 0600)
		defer os.Remove(privateKeyFile)
		if err != nil {
			log.Fatalf("Unable to write sealed-secrets-key to disk: %v", err)
		}
		kubeseal := p.GetBinary("kubeseal")
		flagString := strings.Join(flags, " ")
		argString := strings.Join(args, " ")
		if err := kubeseal(flagString + argString); err != nil {
			log.Fatalf("failed to run kubeseal: %v", err)
		}
	},
}

func init() {
	Unseal.Flags().StringP("format", "o", "yaml", "Output format for sealed secret. Either json or yaml")
	Unseal.Flags().Bool("offline", false, "Decrypt with public and private keys on disk and not from the cluster")
}
