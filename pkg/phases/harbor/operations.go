package harbor

import (
	"fmt"

	"github.com/flanksource/karina/pkg/platform"
	"github.com/pkg/errors"
)

func ReplicateAll(p *platform.Platform) error {
	client, err := NewClient(p)
	if err != nil {
		return err
	}

	p.Infof("Listing replication policies")
	replications, err := client.ListReplicationPolicies()
	if err != nil {
		return fmt.Errorf("replicateAll: failed to list replication policies: %v", err)
	}
	for _, r := range replications {
		p.Infof("Triggering replication of %s (%d)\n", r.Name, r.ID)
		req, err := client.TriggerReplication(r.ID)
		if err != nil {
			return fmt.Errorf("replicateAll: failed to trigger replication: %v", err)
		}
		p.Infof("%s %s: %s  pending: %d, success: %d, failed: %d\n", req.StartTime, req.Status, req.StatusText, req.InProgress, req.Succeed, req.Failed)
	}
	return nil
}

func UpdateSettings(p *platform.Platform) error {
	client, err := NewClient(p)
	if err != nil {
		return err
	}
	p.Infof("Platform: %v", p)
	p.Infof("Settings: %v", *p.Harbor.Settings)
	return client.UpdateSettings(*p.Harbor.Settings)
}

func ListImages(p *platform.Platform, project string, listTags bool) error {
	client, err := NewClient(p)
	if err != nil {
		return errors.Wrap(err, "failed to create harbor client")
	}

	images, err := client.ListImages(project)
	if err != nil {
		return errors.Wrapf(err, "failed to list images for project %s", project)
	}

	for _, image := range images {
		fmt.Printf("Name: %s\n", image.Name)
	}

	return nil
}
