package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/phases/audit"
	"github.com/moshloop/platform-cli/pkg/phases/base"
	"github.com/moshloop/platform-cli/pkg/phases/configmapreloader"
	"github.com/moshloop/platform-cli/pkg/phases/dex"
	"github.com/moshloop/platform-cli/pkg/phases/eck"
	"github.com/moshloop/platform-cli/pkg/phases/fluentdoperator"
	"github.com/moshloop/platform-cli/pkg/phases/flux"
	"github.com/moshloop/platform-cli/pkg/phases/harbor"
	"github.com/moshloop/platform-cli/pkg/phases/monitoring"
	"github.com/moshloop/platform-cli/pkg/phases/nsx"
	"github.com/moshloop/platform-cli/pkg/phases/opa"
	"github.com/moshloop/platform-cli/pkg/phases/postgresoperator"
	"github.com/moshloop/platform-cli/pkg/phases/quack"
	"github.com/moshloop/platform-cli/pkg/phases/registrycreds"
	"github.com/moshloop/platform-cli/pkg/phases/sealedsecrets"
	"github.com/moshloop/platform-cli/pkg/phases/stubs"
	"github.com/moshloop/platform-cli/pkg/phases/vault"
	"github.com/moshloop/platform-cli/pkg/phases/velero"
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	wait                      int
	failOnError               bool
	waitInterval              int
	junitPath, suiteName      string
	thanosURL, pushGatewayURL string
	p                         *platform.Platform
	testAll                   bool
	testDestructive           bool
	testWrite                 bool
)

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
		os.MkdirAll(path.Dir(junitPath), 0755)         // nolint: errcheck
		ioutil.WriteFile(junitPath, []byte(xml), 0644) // nolint: errcheck
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
		elapsed := time.Since(start)
		if test.FailCount == 0 || wait == 0 || int(elapsed.Seconds()) >= wait {
			end(test)
			return
		}
		log.Debugf("Waiting to re-run tests\n")
		time.Sleep(time.Duration(waitInterval) * time.Second)
		log.Infof("Re-running tests\n")
	}
}

func runWithArgs(fn func(p *platform.Platform, test *console.TestResults, args []string, cmd *cobra.Command), args []string, cmd *cobra.Command) {
	start := time.Now()
	for {
		test := console.TestResults{}
		fn(p, &test, args, cmd)
		test.Done()
		elapsed := time.Since(start)
		if test.FailCount == 0 || wait == 0 || int(elapsed.Seconds()) >= wait {
			end(test)
			return
		}
		log.Debugf("Waiting to re-run tests\n")
		time.Sleep(time.Duration(waitInterval) * time.Second)
		log.Infof("Re-running tests\n")
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
			run(fluentdoperator.Test)
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
			run(postgresoperator.Test)
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

	thanosTestCmd := &cobra.Command{
		Use:   "thanos",
		Short: "Test thanos. Requires Pushgateway and Thanos endpoints",
		Long:  "Push metric to pushgateway and try to pull from Thanos. For client cluster --thanos flag is required.",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runWithArgs(monitoring.TestThanos, args, cmd)
		},
	}

	thanosTestCmd.PersistentFlags().StringVarP(&pushGatewayURL, "pushgateway", "p", "", "Url of Pushgateway")
	thanosTestCmd.PersistentFlags().StringVarP(&thanosURL, "thanos", "t", "", "Url of Pushgateway")
	Test.AddCommand(thanosTestCmd)

	Test.AddCommand(&cobra.Command{
		Use:   "sealed-secrets",
		Short: "Test sealed secrets",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(sealedsecrets.Test)
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "vault",
		Short: "Test vault",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(vault.Test)
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "audit",
		Short: "Test kubernetes audit",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(audit.Test)
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "prometheus",
		Short: "Test prometheus",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runWithArgs(monitoring.TestPrometheus, args, cmd)
		},
	})

	configmapReloaderCmd := &cobra.Command{
		Use:   "configmap-reloader",
		Short: "Test configmap-reloader",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runWithArgs(configmapreloader.Test, args, cmd)
		},
	}

	configmapReloaderCmd.PersistentFlags().Bool("e2e", false, "Run e2e tests after main test")
	Test.AddCommand(configmapReloaderCmd)

	Test.AddCommand(&cobra.Command{
		Use:   "gitops",
		Short: "Test gitops",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(flux.Test)
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "quack",
		Short: "Test quack",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(quack.Test)
		},
	})

	Test.AddCommand(&cobra.Command{
		Use:   "registry-credentials",
		Short: "Test registry credentials",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(registrycreds.Test)
		},
	})

	testAllCmd := &cobra.Command{
		Use:   "all",
		Short: "Test all components",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			run(func(p *platform.Platform, test *console.TestResults) {
				client, _ := p.GetClientset()
				base.Test(p, test)
				audit.Test(p, test)

				if testAll || testWrite {
					velero.Test(p, test)
					dex.Test(p, test)
					sealedsecrets.Test(p, test)
					flux.Test(p, test)
					configmapreloader.Test(p, test, args, cmd)
					quack.Test(p, test)
					registrycreds.Test(p, test)
				}
				opa.TestNamespace(p, client, test)
				harbor.Test(p, test)
				monitoring.Test(p, test)
				nsx.Test(p, test)
				fluentdoperator.Test(p, test)
				eck.Test(p, test)
				postgresoperator.Test(p, test)
				vault.Test(p, test)
			})
		},
	}

	testAllCmd.PersistentFlags().BoolVarP(&testWrite, "write", "w", false, "Run write tests")
	testAllCmd.PersistentFlags().BoolVarP(&testDestructive, "destructive", "d", false, "Run destructive tests")
	testAllCmd.PersistentFlags().BoolVarP(&testAll, "all", "a", false, "Run all tests")
	Test.AddCommand(testAllCmd)
}
