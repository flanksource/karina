package velero

import (
	"github.com/moshloop/commons/console"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Test(p *platform.Platform, test *console.TestResults) {
	client, err := p.GetClientset()
	if err != nil {
		test.Failf("velero", "Failed to get k8s client %v", err)
		return
	}

	k8s.TestNamespace(client, Namespace, test)

	if backup, err := CreateBackup(p); err != nil {
		test.Failf("velero", "Failed to create backup: %v", err)
	} else {
		test.Passf("velero", "Backup %s created successfully in %s", backup.Metadata.Name, backup.Status.CompletionTimestamp.Sub(backup.Status.StartTimestamp.Time))
	}

}
