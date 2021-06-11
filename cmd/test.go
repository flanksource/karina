package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/flanksource/karina/pkg/phases/konfigmanager"

	"github.com/flanksource/karina/pkg/phases/mongodboperator"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/karina/pkg/phases/antrea"
	"github.com/flanksource/karina/pkg/phases/apacheds"
	"github.com/flanksource/karina/pkg/phases/argocdoperator"
	"github.com/flanksource/karina/pkg/phases/argorollouts"
	"github.com/flanksource/karina/pkg/phases/base"
	"github.com/flanksource/karina/pkg/phases/calico"
	"github.com/flanksource/karina/pkg/phases/canary"
	"github.com/flanksource/karina/pkg/phases/certmanager"
	"github.com/flanksource/karina/pkg/phases/configmapreloader"
	"github.com/flanksource/karina/pkg/phases/consul"
	"github.com/flanksource/karina/pkg/phases/csi/localpath"
	"github.com/flanksource/karina/pkg/phases/csi/nfs"
	"github.com/flanksource/karina/pkg/phases/csi/s3"
	"github.com/flanksource/karina/pkg/phases/dex"
	"github.com/flanksource/karina/pkg/phases/eck"
	"github.com/flanksource/karina/pkg/phases/elasticsearch"
	"github.com/flanksource/karina/pkg/phases/externaldns"
	"github.com/flanksource/karina/pkg/phases/flux"
	"github.com/flanksource/karina/pkg/phases/gitoperator"
	"github.com/flanksource/karina/pkg/phases/harbor"
	"github.com/flanksource/karina/pkg/phases/istiooperator"
	"github.com/flanksource/karina/pkg/phases/karinaoperator"
	"github.com/flanksource/karina/pkg/phases/keptn"
	"github.com/flanksource/karina/pkg/phases/kiosk"
	"github.com/flanksource/karina/pkg/phases/kpack"
	"github.com/flanksource/karina/pkg/phases/kubeadm"
	"github.com/flanksource/karina/pkg/phases/kuberesourcereport"
	"github.com/flanksource/karina/pkg/phases/kubewebview"
	"github.com/flanksource/karina/pkg/phases/minio"
	"github.com/flanksource/karina/pkg/phases/monitoring"
	"github.com/flanksource/karina/pkg/phases/nginx"
	"github.com/flanksource/karina/pkg/phases/nsx"
	"github.com/flanksource/karina/pkg/phases/opa"
	"github.com/flanksource/karina/pkg/phases/platformoperator"
	"github.com/flanksource/karina/pkg/phases/postgresoperator"
	"github.com/flanksource/karina/pkg/phases/quack"
	"github.com/flanksource/karina/pkg/phases/rabbitmqoperator"
	"github.com/flanksource/karina/pkg/phases/redisoperator"
	"github.com/flanksource/karina/pkg/phases/registrycreds"
	"github.com/flanksource/karina/pkg/phases/sealedsecrets"
	"github.com/flanksource/karina/pkg/phases/templateoperator"
	"github.com/flanksource/karina/pkg/phases/vault"
	"github.com/flanksource/karina/pkg/phases/velero"
	"github.com/flanksource/karina/pkg/phases/vsphere"
	"github.com/flanksource/karina/pkg/platform"

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
	logger.Infof("%s", test)
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
		var writer io.Writer
		writer = &stdout
		if !showProgress {
			writer = os.Stdout
		}
		localTest := &console.TestResults{Writer: writer}
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
		testClient, err := p.GetClientset()
		if err != nil || testClient == nil {
			p.Fatalf("test.go", "Could not establish connection to Platform, aborting tests: %s", err)
			os.Exit(1)
		}
	}

	Test.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		progress.Wait()
		wg.Wait()
		fmt.Println(stdout.String())
		end(test)
	}

	aliases := map[string][]string{
		"bootstrap": []string{"cni", "csi", "base", "cloud", "cert-manager", "nginx", "quack", "minio", "template-operator", "postgres-operator"},
		"cni":       []string{"calico", "antrea", "nsx"},
		"csi":       []string{"s3", "nfs", "local-path"},
		"cloud":     []string{"vsphere"},
		"stubs":     []string{"apacheds", "minio"},
	}
	tests := map[string]TestFn{
		"antrea":               antrea.Test,
		"apacheds":             apacheds.Test,
		"argo-rollouts":        argorollouts.Test,
		"argocd-operator":      argocdoperator.Test,
		"audit":                kubeadm.TestAudit,
		"base":                 base.Test,
		"calico":               calico.Test,
		"canary":               canary.TestCanary,
		"cert-manager":         certmanager.Test,
		"configmap-reloader":   configmapreloader.Test,
		"consul":               consul.Test,
		"dex":                  dex.Test,
		"eck":                  eck.Test,
		"elasticsearch":        elasticsearch.Test,
		"encryption":           kubeadm.TestEncryption,
		"externaldns":          externaldns.Test,
		"git-operator":         gitoperator.Test,
		"gitops":               flux.Test,
		"harbor":               harbor.Test,
		"istio-operator":       istiooperator.Test,
		"keptn":                keptn.Test,
		"karina-operator":      karinaoperator.Test,
		"kiosk":                kiosk.Test,
		"konfig-manager":       konfigmanager.Test,
		"kpack":                kpack.Test,
		"kube-resource-report": kuberesourcereport.TestKubeResourceReport,
		"kube-web-view":        kubewebview.TestKubeWebView,
		"local-path":           localpath.Test,
		"minio":                minio.Test,
		"mongodb-operator":     mongodboperator.Test,
		"monitoring":           monitoring.Test,
		"nfs":                  nfs.Test,
		"nginx":                nginx.Test,
		"nsx":                  nsx.Test,
		"opa":                  opa.Test,
		"platform-operator":    platformoperator.Test,
		"postgres-operator":    postgresoperator.Test,
		"prometheus":           monitoring.TestPrometheus,
		"quack":                quack.Test,
		"rabbitmq-operator":    rabbitmqoperator.Test,
		"redis-operator":       redisoperator.Test,
		"registry-creds":       registrycreds.Test,
		"s3":                   s3.Test,
		"sealed-secrets":       sealedsecrets.Test,
		"template-operator":    templateoperator.Test,
		"thanos":               monitoring.TestThanos,
		"vault":                vault.Test,
		"velero":               velero.Test,
		"vsphere":              vsphere.Test,
	}

	var exec func(name string)

	alreadyRun := make(map[string]bool)

	exec = func(name string) {
		if ok := alreadyRun[name]; ok {
			return
		}
		alreadyRun[name] = true
		if _tests, ok := aliases[name]; ok {
			for _, test := range _tests {
				exec(test)
			}
		} else {
			queue(name, tests[name], wg, ch)
		}
	}
	var Phases = &cobra.Command{
		Use: "phases",
		Run: func(cmd *cobra.Command, args []string) {
			for name := range aliases {
				flag, _ := cmd.Flags().GetBool(name)
				if !flag {
					continue
				}
				exec(name)
			}
			for name := range tests {
				flag, _ := cmd.Flags().GetBool(name)
				if !flag {
					continue
				}
				exec(name)
			}
		},
	}

	Test.AddCommand(Phases)

	for name := range tests {
		_name := name
		Phases.Flags().Bool(name, false, "Test "+name)
		Test.AddCommand(&cobra.Command{
			Use:  name,
			Args: cobra.MinimumNArgs(0),
			Run: func(cmd *cobra.Command, args []string) {
				exec(_name)
			},
		})
	}

	for name, tests := range aliases {
		_name := name
		Phases.Flags().Bool(name, false, fmt.Sprintf("Test %v", tests))
		Test.AddCommand(&cobra.Command{
			Use:  name,
			Args: cobra.MinimumNArgs(0),
			Run: func(cmd *cobra.Command, args []string) {
				exec(_name)
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
