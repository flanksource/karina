package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/flanksource/karina/pkg/phases/canary"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/phases/base"
	"github.com/flanksource/karina/pkg/phases/configmapreloader"
	"github.com/flanksource/karina/pkg/phases/consul"
	"github.com/flanksource/karina/pkg/phases/dex"
	"github.com/flanksource/karina/pkg/phases/eck"
	"github.com/flanksource/karina/pkg/phases/elasticsearch"
	"github.com/flanksource/karina/pkg/phases/fluentdoperator"
	"github.com/flanksource/karina/pkg/phases/flux"
	"github.com/flanksource/karina/pkg/phases/harbor"
	"github.com/flanksource/karina/pkg/phases/kubeadm"
	"github.com/flanksource/karina/pkg/phases/monitoring"
	"github.com/flanksource/karina/pkg/phases/nsx"
	"github.com/flanksource/karina/pkg/phases/opa"
	"github.com/flanksource/karina/pkg/phases/postgresoperator"
	"github.com/flanksource/karina/pkg/phases/quack"
	"github.com/flanksource/karina/pkg/phases/registrycreds"
	"github.com/flanksource/karina/pkg/phases/sealedsecrets"
	"github.com/flanksource/karina/pkg/phases/stubs"
	"github.com/flanksource/karina/pkg/phases/vault"
	"github.com/flanksource/karina/pkg/phases/velero"
	"github.com/flanksource/karina/pkg/platform"
	tests "github.com/flanksource/karina/pkg/test"
	"github.com/spf13/cobra"
	mpb "github.com/vbauerster/mpb/v5"
)

var (
	wait                  int
	failOnError           bool
	waitInterval          int
	junitPath, suiteName  string
	p                     *platform.Platform
	progress              *mpb.Progress
	test                  *console.TestResults
	testE2E, showProgress bool
	concurrency           int
	ch                    chan int
	wg                    *sync.WaitGroup
	stdout                bytes.Buffer
)

var Test = &cobra.Command{
	Use: "test",
}

func end(test *console.TestResults) {
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

type TestFn func(p *platform.Platform, test *console.TestResults)

func queue(name string, fn TestFn, wg *sync.WaitGroup, ch chan int) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		localTest := &console.TestResults{Writer: &stdout}
		var bar console.Progress
		if showProgress {
			bar = console.NewTerminalProgress(name, localTest, progress)
		} else {
			bar = console.NewTextProgress(name, localTest)
		}
		start := time.Now()
		bar.Start()
		for {
			ch <- 1
			fn(p.WithLogOutput(localTest.Writer).WithField("test", name), localTest)
			elapsed := time.Since(start)
			if localTest.FailCount == 0 || wait == 0 || int(elapsed.Seconds()) >= wait {
				test.Append(localTest)
				bar.Done()
				<-ch
				return
			}
			localTest.Retry()
			bar.Status(fmt.Sprintf("Waiting to re-run tests, retry: %d", localTest.Retries))
			time.Sleep(time.Duration(waitInterval) * time.Second)
			bar.Status("Re-running tests")
			<-ch
		}
	}()
}

func init() {
	Test.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		GlobalPreRun(cmd, args)
		p = getPlatform(cmd)
		wg = &sync.WaitGroup{}
		ch = make(chan int, concurrency)
		progress = mpb.New(mpb.WithWaitGroup(wg), mpb.WithWidth(40))
		test = &console.TestResults{
			Writer: &stdout,
		}
	}
	Test.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		progress.Wait()
		wg.Wait()
		fmt.Println(stdout.String())
		end(test)
	}

	tests := map[string]TestFn{
		"audit":              kubeadm.TestAudit,
		"base":               base.Test,
		"canary":             canary.TestCanary,
		"configmap-reloader": configmapreloader.Test,
		"consul":             consul.Test,
		"dex":                dex.Test,
		"eck":                eck.Test,
		"elasticsearch":      elasticsearch.Test,
		"encryption":         kubeadm.TestEncryption,
		"fluentd":            fluentdoperator.Test,
		"gitops":             flux.Test,
		"harbor":             harbor.Test,
		"kube-web-view":      monitoring.TestKubeWebView,
		"monitoring":         monitoring.Test,
		"nsx":                nsx.Test,
		"opa":                opa.Test,
		"postgres-operator":  postgresoperator.Test,
		"promtheus":          monitoring.TestPrometheus,
		"quack":              quack.Test,
		"registry-creds":     registrycreds.Test,
		"sealed-secrets":     sealedsecrets.Test,
		"stubs":              stubs.Test,
		"templates":          tests.TestTemplates,
		"thanos":             monitoring.TestThanos,
		"vault":              vault.Test,
		"velero":             velero.Test,
	}

	var Phases = &cobra.Command{
		Use: "phases",
		Run: func(cmd *cobra.Command, args []string) {
			for name, fn := range tests {
				_name := name
				_fn := fn
				flag, _ := cmd.Flags().GetBool(name)
				if !flag {
					continue
				}
				queue(_name, _fn, wg, ch)
			}
		},
	}

	Test.AddCommand(Phases)

	for name, fn := range tests {
		_name := name
		_fn := fn
		Phases.Flags().Bool(name, false, "Test "+name)
		Test.AddCommand(&cobra.Command{
			Use:  name,
			Args: cobra.MinimumNArgs(0),
			Run: func(cmd *cobra.Command, args []string) {
				queue(_name, _fn, wg, ch)
			},
		})
	}

	testAllCmd := &cobra.Command{
		Use:   "all",
		Short: "Test all components",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			for name, fn := range tests {
				if Contains(p.Test.Exclude, name) {
					test.Skipf(name, name)
					continue
				}
				queue(name, fn, wg, ch)
			}
		},
	}
	Test.PersistentFlags().IntVar(&wait, "wait", 0, "Time in seconds to wait for tests to pass")
	Test.PersistentFlags().IntVar(&waitInterval, "wait-interval", 5, "Time in seconds to wait between repeated tests")
	Test.PersistentFlags().StringVar(&junitPath, "junit-path", "", "Path to export JUnit formatted test results")
	Test.PersistentFlags().StringVar(&suiteName, "suite-name", "", "Name of the Test Suite, defaults to platform name")
	Test.PersistentFlags().BoolVar(&failOnError, "fail-on-error", true, "Return an exit code of 1 if any tests fail")
	Test.PersistentFlags().BoolVarP(&testE2E, "e2e", "", false, "Run e2e tests")
	Test.PersistentFlags().BoolVar(&showProgress, "progress", true, "Display progress as tests run")
	Test.PersistentFlags().IntVar(&concurrency, "concurrency", 8, "Number of tests to run concurrently")
	Test.AddCommand(testAllCmd)
}

// Contains tells whether a contains x.
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
