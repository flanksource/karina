package ingress

import (
	"github.com/flanksource/karina/pkg/phases/nginx"
	"github.com/flanksource/karina/pkg/platform"
)

func Install(p *platform.Platform) error {
	return nginx.Install(p)
}
