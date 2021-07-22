package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func getFlagString(name string, cmd *cobra.Command) string {
	value, err := cmd.Flags().GetString(name)
	if value == "" || err != nil {
		return ""
	}
	return fmt.Sprintf("--%s %s", name, value)
}

func getFlagFilePath(name string, cmd *cobra.Command) string {
	value, err := cmd.Flags().GetString(name)
	if value == "" || err != nil {
		return ""
	}
	filePath, err := filepath.Abs(value)
	if err == nil {
		log.Warningf("Unable to get absolute path for %s: %v", value, err)
		return ""
	}
	return fmt.Sprintf("--%s %s", name, filePath)
}

func getFlagFilePathSlice(name string, cmd *cobra.Command) string {
	argList, err := cmd.Flags().GetStringSlice(name)
	if err != nil || len(argList) == 0 {
		return ""
	}
	result := ""
	for _, value := range argList {
		filePath, err := filepath.Abs(value)
		if err == nil {
			log.Warningf("Unable to get absolute path for %s: %v", value, err)
		} else {
			result = fmt.Sprintf("%s --%s %s", result, name, filePath)
		}
	}
	return fmt.Sprintf("--%s %s", name, result)
}

func getFlagBool(name string, cmd *cobra.Command) string {
	value, _ := cmd.Flags().GetBool(name)
	if !value {
		return ""
	}
	return fmt.Sprintf("--%s", name)
}

var Seal = &cobra.Command{
	Use:   "seal",
	Short: "Seal a secret using sealed-secrets",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		p := getPlatform(cmd)
		flags := []string{
			getFlagString("format", cmd),
			getFlagFilePath("merge-into", cmd),
			getFlagBool("re-encrypt", cmd),
			getFlagString("name", cmd),
			getFlagFilePathSlice("from-file", cmd),
		}

		if !p.SealedSecrets.Disabled && p.SealedSecrets.Certificate != nil {
			if p.SealedSecrets.Certificate.Cert == "" {
				log.Fatalf("Sealed-secrets certificate not provided in config")
			}
			flags = append(flags, "--cert", p.SealedSecrets.Certificate.Cert)
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
	Seal.Flags().StringP("format", "o", "yaml", "Output format for sealed secret. Either json or yaml")
	Seal.Flags().String("merge-into", "", "Merge items from secret into an existing sealed secret file, updating the file in-place instead of writing to stdout.")
	Seal.Flags().Bool("re-encrypt", false, "Re-encrypt the given sealed secret to use the latest cluster key.")
	Seal.Flags().String("name", "", "Name of the sealed secret (required with --raw and default (strict) scope)")
	Seal.Flags().StringSlice("from-file", nil, "(only with --raw) Secret items can be sourced from files. Pro-tip: you can use /dev/stdin to read pipe input. This flag tries to follow the same syntax as in kubectl")
}
