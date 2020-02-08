package harbor

import (
	"github.com/moshloop/platform-cli/pkg/phases/pgo"
	"github.com/moshloop/platform-cli/pkg/platform"
	log "github.com/sirupsen/logrus"
)

func Backup(p *platform.Platform) error {
	for _, db := range dbNames {
		if err := pgo.Backup(p, dbCluster, db); err != nil {
			log.Tracef("Backup: Failed to create backup: %s", err)
			return err
		}
	}
	return nil
}

func Restore(p *platform.Platform, backup string) error {
	for _, db := range dbNames {
		if err := pgo.Restore(p, dbCluster, db); err != nil {
			log.Tracef("Backup: Failed to restore backup: %s", err)
			return err
		}
	}
	return nil
}
