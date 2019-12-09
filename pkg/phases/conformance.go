package phases

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/platform"
)

func ConformanceTest(p *platform.Platform, options ConformanceTestOptions) error {

	sonobuoy := p.GetBinaryWithKubeConfig("sonobuoy")
	args := ""
	args += fmt.Sprintf(" --wait=%d --wait-output Spinner", options.Wait)

	if options.Quick {
		args += " --mode quick"
	}

	if log.IsLevelEnabled(log.DebugLevel) {
		args += " -v 1"
	} else if log.IsLevelEnabled(log.TraceLevel) {
		args += " -v 2"
	}

	if err := sonobuoy("run %s", args); err != nil {
		return fmt.Errorf("Error submitting sonobuoy tests: %v", err)
	}
	if err := sonobuoy("retrieve %s", options.OutputDir); err != nil {
		return fmt.Errorf("Error retrieving sonobuoy test results: %v", err)
	}
	return sonobuoy("delete --all")
}

type ConformanceTestOptions struct {
	Certification, KubeBench, Quick bool
	Wait                            int
	OutputDir                       string
}
