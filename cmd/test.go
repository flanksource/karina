package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/flanksource/commons/console"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/moshloop/platform-cli/pkg/phases/base"
	"github.com/moshloop/platform-cli/pkg/phases/contour"
	"github.com/moshloop/platform-cli/pkg/phases/dex"
	"github.com/moshloop/platform-cli/pkg/phases/eck"
	"github.com/moshloop/platform-cli/pkg/phases/fluentdOperator"
	"github.com/moshloop/platform-cli/pkg/phases/harbor"
	"github.com/moshloop/platform-cli/pkg/phases/monitoring"
	"github.com/moshloop/platform-cli/pkg/phases/nsx"
	"github.com/moshloop/platform-cli/pkg/phases/opa"
	"github.com/moshloop/platform-cli/pkg/phases/postgresOperator"
	"github.com/moshloop/platform-cli/pkg/phases/stubs"
	"github.com/moshloop/platform-cli/pkg/phases/velero"
	"github.com/moshloop/platform-cli/pkg/platform"
)

var wait int
var failOnError bool
var waitInterval int
var junitPath, suiteName string
var p *platform.Platform

var Test = &cobra.Command{
	Use: "test",
}

func end(test console.TestResults) {
	if junitPath != "" {
		if suiteName == "" {
			test.SuiteName(p.Name)
		} else {
			test.SuiteName(suiteName)
		}
		xml, _ := test.ToXML()
		os.MkdirAll(path.Dir(junitPath), 0755)
		ioutil.WriteFile(junitPath, []byte(xml), 0644)
	}
	if test.FailCount > 0 && failOnError {
		os.Exit(1)
	}
}

func run(fn func(p *platform.Platform, test *console.TestResults)) {
	start := time.Now()
	for {
		test := console.TestResults{}
		fn(p, &test)
		test.Done()
		elapsed := time.Now().Sub(start)
		if test.FailCount == 0 || wait == 0 || int(elapsed.Seconds()) >= wait {
			end(test)
			return
		} else {
			log.Debugf("Waiting to re-run tests\n")
			time.Sleep(time.Duration(waitInterval) * time.Second)
			log.Infof("Re-running tests\n")

		}
	}
}

func init() {
	Test.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		p = getPlatform(cmd)
	}
	Test.PersistentFlags().IntVar(&wait, "wait", 0, "Time in seconds to wait for tests to pass")
	Test.PersistentFlags().IntVar(&waitInterval, "wait-interval", 5, "Time in seconds to wait between repeated tests")
	Test.PersistentFlags().StringVar(&junitPath, "junit-path", "", "Path to export JUnit formatted test results")
	Test.PersistentFlags().StringVar(&suiteName, "suite-name", "", "Name of the Test Suite, defaults to platform name")
	Test.PersistentFlags().BoolVar(&failOnError, "fail-on-error", true, "Return an exit code of 1 if any tests fail")
	Test.AddCommand(&cobra.Command{
		Use:   "harbor",
		Short: "Test harbor",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(harbor.Test)
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "dex",
		Short: "Test dex",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(dex.Test)
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "opa",
		Short: "Test opa policies using a fixtures director",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			run(func(p *platform.Platform, test *console.TestResults) {
				opa.TestPolicies(p, args[0], test)
			})
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "base",
		Short: "Test base",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(base.Test)
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "nsx",
		Short: "Test NSX-T CNI",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(nsx.Test)
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "velero",
		Short: "Test velero",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(velero.Test)
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "monitoring",
		Short: "Test monitoring stack",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(monitoring.Test)
		},
	})
	Test.AddCommand(&cobra.Command{
		Use:   "fluentd",
		Short: "Test fluentd",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(fluentdOperator.Test)
		},
	})
	Test.AddCommand(&cobra.Command{
		Use:   "eck",
		Short: "Test eck",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(eck.Test)
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "postgres-operator",
		Short: "Test postgres-operator",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(postgresOperator.Test)
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "stubs",
		Short: "Test stubs",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(stubs.Test)
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "contour",
		Short: "Test contour",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(contour.Test)
		},
	})
	Test.AddCommand(&cobra.Command{
		Use:   "all",
		Short: "Test all components",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(func(p *platform.Platform, test *console.TestResults) {
				client, _ := p.GetClientset()
				base.Test(p, test)
				velero.Test(p, test)
				opa.TestNamespace(p, client, test)
				harbor.Test(p, test)
				dex.Test(p, test)
				monitoring.Test(p, test)
				nsx.Test(p, test)
				fluentdOperator.Test(p, test)
				eck.Test(p, test)
				contour.Test(p, test)
				postgresOperator.Test(p, test)
			})
		},
	})
}
