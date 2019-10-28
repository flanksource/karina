package harbor

import (
	"github.com/moshloop/platform-cli/pkg/phases/pgo"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func Backup(p *platform.Platform) error {
	for _, db := range dbNames {
		if err := pgo.Backup(p, dbCluster, db); err != nil {
			return err
		}
	}
	return nil
}

func Restore(p *platform.Platform, backup string) error {
	for _, db := range dbNames {
		if err := pgo.Restore(p, dbCluster, db); err != nil {
			return err
		}
	}
	return nil
}
