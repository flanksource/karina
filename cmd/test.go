package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/moshloop/commons/console"
	"github.com/moshloop/platform-cli/pkg/phases/base"
	"github.com/moshloop/platform-cli/pkg/phases/dex"
	"github.com/moshloop/platform-cli/pkg/phases/harbor"
	"github.com/moshloop/platform-cli/pkg/phases/monitoring"
	"github.com/moshloop/platform-cli/pkg/phases/pgo"
)

var Test = &cobra.Command{
	Use: "test",
}

func init() {
	Test.AddCommand(&cobra.Command{
		Use:   "harbor",
		Short: "Test harbor installation",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			test := console.TestResults{}
			harbor.Test(getPlatform(cmd), &test)
			test.Done()
			if test.FailCount > 0 {
				os.Exit(1)
			}
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "dex",
		Short: "Test dex ",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			test := console.TestResults{}
			dex.Test(getPlatform(cmd), &test)
			test.Done()
			if test.FailCount > 0 {
				os.Exit(1)
			}
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "pgo",
		Short: "Test postgres operator ",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			test := console.TestResults{}
			pgo.Test(getPlatform(cmd), &test)
			test.Done()
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "base",
		Short: "Test base installation",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			test := console.TestResults{}
			pgo.Test(getPlatform(cmd), &test)
			test.Done()
			if test.FailCount > 0 {
				os.Exit(1)
			}
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "monitoring",
		Short: "Test monitoring stack installation",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			test := console.TestResults{}
			monitoring.Test(getPlatform(cmd), &test)
			test.Done()
			if test.FailCount > 0 {
				os.Exit(1)
			}
		},
	})
	Test.AddCommand(&cobra.Command{
		Use:   "all",
		Short: "Test all components",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			test := console.TestResults{}
			p := getPlatform(cmd)
			base.Test(p, &test)
			pgo.Test(p, &test)
			harbor.Test(p, &test)
			dex.Test(p, &test)
			test.Done()
			if test.FailCount > 0 {
				os.Exit(1)
			}
		},
	})
}
