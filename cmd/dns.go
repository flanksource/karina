package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var DNS = &cobra.Command{
	Use:   "dns",
	Short: "DNS sync",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		dns := getPlatform(cmd).GetDNSClient()

		if err := dns.Append("k8s.", "test2.k8s.", "127.0.0.1"); err != nil {
			fmt.Println(err)
		}
		// fmt.Println(dns.Get("test.k8s."))

	},
}
