package harbor

import (
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
)

func ReplicateAll(p *platform.Platform) error {
	defaults(p)
	client := NewHarborClient(p)

	log.Infoln("Listing replication policies")
	replications, err := client.ListReplicationPolicies()
	if err != nil {
		return err
	}
	for _, r := range replications {
		log.Infof("Triggering replication of %s (%d)\n", r.Name, r.ID)
		req, err := client.TriggerReplication(r.ID)
		if err != nil {
			return err
		}
		log.Infof("%s %s: %s  pending: %d, success: %d, failed: %d\n", req.StartTime, req.Status, req.StatusText, req.InProgress, req.Succeed, req.Failed)
	}
	return nil
}
