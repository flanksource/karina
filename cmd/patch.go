package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	deploy_base "github.com/moshloop/platform-cli/pkg/phases/base"
	"github.com/moshloop/platform-cli/pkg/phases/calico"
	"github.com/moshloop/platform-cli/pkg/phases/dex"
	"github.com/moshloop/platform-cli/pkg/phases/flux"
	"github.com/moshloop/platform-cli/pkg/phases/harbor"
	"github.com/moshloop/platform-cli/pkg/phases/monitoring"
	"github.com/moshloop/platform-cli/pkg/phases/nsx"
	"github.com/moshloop/platform-cli/pkg/phases/opa"
	"github.com/moshloop/platform-cli/pkg/phases/pgo"
	"github.com/moshloop/platform-cli/pkg/phases/stubs"
	"github.com/moshloop/platform-cli/pkg/phases/velero"
)

var Patch = &cobra.Command{
	Use:   "patch",
	Short: "Patch the platform",
}