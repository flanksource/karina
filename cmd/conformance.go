package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/moshloop/platform-cli/pkg/phases"
)

var Conformance = &cobra.Command{
	Use:   "conformance",
	Short: "Run conformance tests using sonobuoy",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := phases.ConformanceTest(getPlatform(cmd), phases.ConformanceTestOptions{
			Certification: certification,
			KubeBench:     kubeBench,
			Quick:         quick,
			Wait:          wait,
			OutputDir:     outputDir,
		}); err != nil {
			log.Fatalf("Failed to run conformance tests: %v", err)
		}
	},
}

var certification, kubeBench, quick bool

// var outputDir string

func init() {
	Conformance.Flags().BoolVar(&certification, "certification", false, "Run full certification tests")
	Conformance.Flags().BoolVar(&kubeBench, "kube-bench", false, "Run kube-bench")
	Conformance.Flags().BoolVar(&quick, "quick", false, "Run quick tests only")
	Conformance.Flags().IntVar(&wait, "wait", 7200, "Wait for tests to complete")
	Conformance.Flags().StringVar(&outputDir, "output-dir", "conformance-tests", "Output directory for conformance tests")
}
