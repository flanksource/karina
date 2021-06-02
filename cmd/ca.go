package cmd

import (
	"fmt"

	"github.com/flanksource/karina/pkg/ca"
	"github.com/flanksource/karina/pkg/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CA = &cobra.Command{
	Use:   "ca",
	Short: "Commands for generating CA certs",
}

var generateCA = &cobra.Command{
	Use:   "generate",
	Short: "Generate CA certificates",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		certPath, _ := cmd.Flags().GetString("cert-path")
		privateKeyPath, _ := cmd.Flags().GetString("private-key-path")
		password, _ := cmd.Flags().GetString("password")
		expiry, _ := cmd.Flags().GetInt("expiry")
		if err := ca.GenerateCA(name, certPath, privateKeyPath, password, expiry); err != nil {
			log.Fatalf("Failed to generate certificate, %s", err)
		}
	},
}

var decrypt = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt a password protected private key",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		privateKeyPath, _ := cmd.Flags().GetString("key")
		password, _ := cmd.Flags().GetString("password")
		certPath, _ := cmd.Flags().GetString("cert")
		cert, err := ca.ReadCA(&types.CA{
			Cert:       certPath,
			PrivateKey: privateKeyPath,
			Password:   password,
		})
		if err != nil {
			log.Fatalf("Failed to decrypt key %s", err)
		}
		fmt.Println(string(cert.EncodedPrivateKey()))
	},
}

var validateCA = &cobra.Command{
	Use:   "validate",
	Short: "Validate CA certificates",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		certPath, _ := cmd.Flags().GetString("cert-path")
		privateKeyPath, _ := cmd.Flags().GetString("private-key-path")
		password, _ := cmd.Flags().GetString("password")
		if err := ca.ValidateCA(certPath, privateKeyPath, password); err != nil {
			log.Fatalf("Failed to validate certificate, %s", err)
		}
	},
}

func init() {
	CA.AddCommand(generateCA, validateCA, decrypt)
	generateCA.Flags().String("name", "", "certificate name")
	generateCA.Flags().String("cert-path", "", "path to certificate file")
	generateCA.Flags().String("private-key-path", "", "path to private key file")
	generateCA.Flags().String("password", "", "certificate password")
	generateCA.Flags().Int("expiry", 1, "certificate expiration in years")
	validateCA.Flags().String("cert-path", "", "path to certificate file")
	validateCA.Flags().String("private-key-path", "", "path to private key file")
	validateCA.Flags().String("password", "", "certificate password")
	decrypt.Flags().String("key", "", "path to private key file")
	decrypt.Flags().String("password", "", "certificate password")
	decrypt.Flags().String("cert", "", "path to certificate file")
}
