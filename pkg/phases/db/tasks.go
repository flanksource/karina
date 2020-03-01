package db

import (
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
)

func GetOrCreateDB(p *platform.Platform, clusterName string, dbNames ...string) (*types.DB, error) {
	return p.GetOrCreateDB(clusterName, dbNames...)
}

func Backup(p *platform.Platform, clusterName string) error {
	return nil
}

func Restore(p *platform.Platform, backup string, clusterName string) error {
	return nil
}
