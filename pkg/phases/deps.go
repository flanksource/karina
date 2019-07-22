package phases

import (
	"os"
	"runtime"

	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/moshloop/platform-cli/pkg/utils"
	log "github.com/sirupsen/logrus"
)

var dependency_versions = map[string]string{
	"gomplate":     "v3.5.0",
	"konfigadm":    "v0.3.6",
	"jb":           "v0.1.0",
	"jsonnet":      "v0.13.0",
	"sonobuoy":     "0.15.0",
	"govc":         "v0.20.0",
	"gojsontoyaml": "master",
	"kustomize":    "v3.0.2",
}
var dependencies_macosx = map[string]string{
	"gomplate":  "https://github.com/hairyhenderson/gomplate/releases/download/{{.version}}/gomplate_darwin-amd64",
	"konfigadm": "https://github.com/moshloop/konfigadm/releases/download/{{.version}}/konfigadm_osx",
	"jb":        "https://github.com/jsonnet-bundler/jsonnet-bundler/releases/download/{{.version}}/jb-darwin-amd64",
	"sonobuoy":  "https://github.com/heptio/sonobuoy/releases/download/v{{.version}}/sonobuoy_{{.version}}_darwin_amd64.tar.gz",
	"govc":      "https://github.com/vmware/govmomi/releases/download/{{.version}}/govc_darwin_amd64.gz",
}

var dependencies_linux = map[string]string{
	"gomplate":  "https://github.com/hairyhenderson/gomplate/releases/download/{{.version}}/gomplate_linux-amd64",
	"konfigadm": "https://github.com/moshloop/konfigadm/releases/download/{{.version}}/konfigadm",
	"jb":        "https://github.com/jsonnet-bundler/jsonnet-bundler/releases/download/{{.version}}/jb-linux-amd64",
	"sonobuoy":  "https://github.com/heptio/sonobuoy/releases/download/v{.version}}/sonobuoy_{{.version}}_linux_amd64.tar.gz",
	"govc":      "https://github.com/vmware/govmomi/releases/download/{{.version}}/govc_linux_amd64.gz",
}

var dependencies_go = map[string]string{
	"jsonnet":      "github.com/google/go-jsonnet/cmd/jsonnet@{{.version}}",
	"gojsontoyaml": "github.com/brancz/gojsontoyaml",
	"kustomize":    "sigs.k8s.io/kustomize/v3/cmd/kustomize",
}

func Deps(cfg types.PlatformConfig) error {
	os.Mkdir(".bin", 0755)
	binaries := dependencies_linux
	if runtime.GOOS == "darwin" {
		binaries = dependencies_macosx
	}

	for dep, ver := range dependency_versions {
		bin := ".bin/" + dep
		if utils.FileExists(bin) {
			log.Debugf("%s already exists", bin)
			continue
		}
		if path, ok := binaries[dep]; ok {
			url := utils.Interpolate(path, map[string]string{"version": ver})
			log.Infof("Installing %s (%s) -> %s", dep, ver, url)
			err := utils.Download(url, bin)
			if err != nil {
				return err
			}
			if err := os.Chmod(bin, 0755); err != nil {
				return err
			}
		} else if gopath, ok := dependencies_go[dep]; ok {
			url := utils.Interpolate(gopath, map[string]string{"version": ver})
			log.Infof("Installing via go get %s (%s) -> %s", dep, ver, url)
			if err := utils.Exec("GOPATH=$PWD/.go go get %s", url); err != nil {
				return err
			}
			if err := os.Rename(".go/bin/"+dep, bin); err != nil {
				return err
			}
		}
	}
	return nil
}
