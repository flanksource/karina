package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var DNS = &cobra.Command{
	Use: "dns",
}

var domain string

func init() {
	append := &cobra.Command{
		Use:  "append",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dns := getPlatform(cmd).GetDNSClient()
			if err := dns.Append(domain, args...); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

		},
	}
	update := &cobra.Command{
		Use:  "update",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dns := getPlatform(cmd).GetDNSClient()
			if err := dns.Update(domain, args...); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

		},
	}

	delete := &cobra.Command{
		Use:  "delete",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dns := getPlatform(cmd).GetDNSClient()
			if err := dns.Delete(domain, args...); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	DNS.AddCommand(append, update, delete)
	DNS.PersistentFlags().StringVar(&domain, "domain", "", "The DNS sub-domain (excluding root Zone) to update")
}
