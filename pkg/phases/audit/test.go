package audit

import (
	"github.com/flanksource/commons/console"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	test.Passf("PassTest","Hello %v", "test")
	test.Failf("FailTest","Hello %v", "test")

}
