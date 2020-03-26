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

	list := &cobra.Command{
		Use:  "list",
		Args: cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			dns := getPlatform(cmd).GetDNSClient()
			list, err := dns.Get(domain)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			for _, record := range list {
				fmt.Println(record)
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
	DNS.AddCommand(append, update, delete, list)
	DNS.PersistentFlags().StringVar(&domain, "domain", "", "The DNS sub-domain (excluding root Zone) to update")
}
