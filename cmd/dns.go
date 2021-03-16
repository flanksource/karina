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

	get := &cobra.Command{
		Use:  "get",
		Args: cobra.MaximumNArgs(0),
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
	DNS.AddCommand(append, update, delete, get)
	DNS.PersistentFlags().StringVar(&domain, "domain", "", "")
	DNS.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if domain == "" {
			fmt.Println("Must specify a --domain")
			os.Exit(1)
		}
	}
	_ = DNS.MarkFlagRequired("domain")
}
