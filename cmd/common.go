package cmd

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/moshloop/commons/is"
	"github.com/moshloop/commons/text"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
)

func getPlatform(cmd *cobra.Command) *platform.Platform {
	platform := platform.Platform{
		PlatformConfig: getConfig(cmd),
	}
	return &platform
}

func getConfig(cmd *cobra.Command) types.PlatformConfig {
	paths, _ := cmd.Flags().GetStringArray("config")
	splitPaths := []string{}
	for _, path := range paths {
		splitPaths = append(splitPaths, strings.Split(path, ",")...)
	}

	if len(paths) == 0 {
		log.Fatalf("Must specify at least 1 config")
	}
	paths = splitPaths
	base := types.PlatformConfig{
		Source: paths[0],
	}

	for _, path := range paths {
		cfg := types.PlatformConfig{
			Source: paths[0],
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("Failed to read config file %s, %s", path, err)
		}

		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("Failed to parse YAML: %s", err)
		}

		for node, vm := range cfg.Nodes {
			if baseNode, ok := base.Nodes[node]; ok {
				if err := mergo.Merge(&baseNode, vm); err != nil {
					log.Fatalf("Failed to merge nodes %s, %s", node, err)
				}
				base.Nodes[node] = baseNode
			}
		}

		if err := mergo.Merge(&base, cfg); err != nil {
			log.Fatalf("Failed to merge in %s, %s", path, err)
		}
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		base.DryRun = true
		log.Infof("Running a dry-run mode, no changes will be made")
	}

	base.S3.AccessKey = template(base.S3.AccessKey)
	base.S3.SecretKey = template(base.S3.SecretKey)

	ldap := base.Ldap
	ldap.Username = template(ldap.Username)
	ldap.Password = template(ldap.Password)
	base.Ldap = ldap

	if base.TrustedCA != "" && !is.File(base.TrustedCA) {
		base.TrustedCA = text.ToFile(base.TrustedCA, ".pem")
	}

	data, _ := yaml.Marshal(base)
	log.Tracef("Using configuration: \n%s\n", string(data))
	base.Init()
	return base
}

func template(val string) string {
	if strings.HasPrefix(val, "$") {
		env := os.Getenv(val[1:])
		if env != "" {
			return env
		}
	}
	return val
}
