package cmd

import (
	"github.com/moshloop/platform-cli/pkg/ca"
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

var validateCA = &cobra.Command{
	Use:   "validate",
	Short: "Validate CA certificates",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		certPath, _ := cmd.Flags().GetString("cert-path")
		privateKeyPath, _ := cmd.Flags().GetString("private-key-path")
		password, _ := cmd.Flags().GetString("password")
		expiry, _ := cmd.Flags().GetInt("expiry")
		if err := ca.ValidateCA(name, certPath, privateKeyPath, password, expiry); err != nil {
			log.Fatalf("Failed to validate certificate, %s", err)
		}
	},
}

func init() {
	CA.AddCommand(generateCA, validateCA)
	generateCA.Flags().String("name", "", "certificate name")
	generateCA.Flags().String("cert-path", "", "path to certificate file")
	generateCA.Flags().String("private-key-path", "", "path to private key file")
	generateCA.Flags().String("password", "", "certificate password")
	generateCA.Flags().String("expiry", "1", "certificate expiration in years")
	validateCA.Flags().String("name", "", "certificate name")
	validateCA.Flags().String("cert-path", "", "path to certificate file")
	validateCA.Flags().String("private-key-path", "", "path to private key file")
	validateCA.Flags().String("password", "", "certificate password")
	validateCA.Flags().String("expiry", "1", "certificate expiration in years")
}
