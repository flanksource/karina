package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	sealedSecretsNamespace = "sealed-secrets"
	privateKeyFile         = ".kubeseal.sealed-secrets-key.json"
)

var Unseal = &cobra.Command{
	Use:   "unseal",
	Short: "Unseal a secret using sealed-secrets",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		p := getPlatform(cmd)
		workingDirectory, err := os.Getwd()
		if err != nil {
			log.Fatalf("Unable to get the current working directory: %v", err)
		}
		flags := []string{
			"--recovery-unseal",
			getFlagString("format", cmd),
			getFlagBool("allow-empty-data", cmd),
			fmt.Sprintf("--recovery-private-key %s/%s", workingDirectory, privateKeyFile),
		}
		if !p.SealedSecrets.Disabled && p.SealedSecrets.Certificate != nil {
			if p.SealedSecrets.Certificate.Cert != "" {
				flags = append(flags, "--cert", p.SealedSecrets.Certificate.Cert)
			}
		}
		client, err := p.GetClientset()
		if err != nil {
			log.Fatalf("Unable to get Kubernetes client set: %v", err)
		}
		secretsInterface := client.CoreV1().Secrets(sealedSecretsNamespace)
		secrets, err := secretsInterface.List(context.TODO(), metav1.ListOptions{
			LabelSelector: "sealedsecrets.bitnami.com/sealed-secrets-key=active",
		})
		if err != nil {
			log.Fatalf("Unable to get sealed-secrets active private key: %v", err)
		}
		if len(secrets.Items) < 1 {
			log.Fatalf("No sealed-secrrts active private key set")
		}
		secret := secrets.Items[0]
		// Required because k8s API client does not set it automatically.
		typeMeta := metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		}
		secret.TypeMeta = typeMeta
		file, err := json.Marshal(secret)
		if err != nil {
			log.Fatalf("Unable to get sealed-secrets active private key: %v", err)
		}
		err = ioutil.WriteFile(privateKeyFile, file, 0600)
		if err != nil {
			log.Fatalf("Unable to write sealed-secrets-key to disk: %v", err)
		}
		kubeseal := p.GetBinary("kubeseal")
		flagString := strings.Join(flags, " ")
		argString := strings.Join(args, " ")
		if err := kubeseal(flagString + argString); err != nil {
			os.Remove(privateKeyFile)
			log.Fatalf("failed to run kubeseal: %v", err)
		}
		os.Remove(privateKeyFile)
	},
}

func init() {
	Unseal.Flags().StringP("format", "o", "json", "Output format for sealed secret. Either json or yaml")
	Unseal.Flags().Bool("allow-empty-data", false, "Allow empty data in the secret object")
}
