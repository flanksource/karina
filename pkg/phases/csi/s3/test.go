package s3

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if !p.S3.CSIVolumes {
		return
	}
	client, _ := p.GetClientset()
	kommons.TestDeploy(client, Namespace, "csi-attacher-s3", test)
}
