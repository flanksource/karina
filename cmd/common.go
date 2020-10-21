package cmd

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/is"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/lookup"
	"github.com/flanksource/commons/text"
	"github.com/flanksource/karina/pkg/constants"
	"github.com/flanksource/karina/pkg/phases/harbor"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/karina/pkg/types"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	yaml "gopkg.in/flanksource/yaml.v3"
)

func getPlatform(cmd *cobra.Command) *platform.Platform {
	platform := platform.Platform{
		PlatformConfig: getConfig(cmd),
	}
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	if kubeconfig != "" {
		platform.KubeConfigPath = kubeconfig
	}
	if err := platform.Init(); err != nil {
		log.Fatalf("failed to initialise: %v", err)
	}
	return &platform
}

func getConfig(cmd *cobra.Command) types.PlatformConfig {
	paths, _ := cmd.Flags().GetStringArray("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	extras, _ := cmd.Flags().GetStringArray("extra")
	trace, _ := cmd.Flags().GetBool("trace")
	e2e, _ := cmd.Flags().GetBool("e2e")

	config := NewConfig(paths, extras)
	config.E2E = e2e
	config.Trace = trace
	config.DryRun = dryRun
	return config
}

func NewConfig(paths []string, extras []string) types.PlatformConfig {
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

	if err := mergeConfigs(&base, paths); err != nil {
		log.Fatalf("Failed to merge configs: %v", err)
	}

	defaultConfig := types.DefaultPlatformConfig()
	if err := mergo.Merge(&base, defaultConfig); err != nil {
		log.Fatalf("Failed to merge default config, %v", err)
	}

	ldap := base.Ldap
	if ldap.Port == "" {
		ldap.Port = "636"
	}
	if ldap != nil {
		base.Ldap = ldap
	}

	if base.TrustedCA != "" && !is.File(base.TrustedCA) {
		base.TrustedCA = text.ToFile(base.TrustedCA, ".pem")
	}

	base.Gatekeeper.WhitelistNamespaces = append(base.Gatekeeper.WhitelistNamespaces, constants.PlatformNamespaces...)

	for _, extra := range extras {
		key := strings.Split(extra, "=")[0]
		val := extra[len(key)+1:]
		log.Debugf("Looking up %s to set it to: %s", key, val)

		value, err := lookup.LookupString(&base, key)
		if err != nil {
			log.Fatalf("Cannot lookup %s: %v", key, err)
		}
		log.Infof("Overriding %s %v => %v", key, value, val)
		switch value.Interface().(type) {
		case string:
			value.SetString(val)
		case int:
			i, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				log.Fatalf("Cannot convert %s to int", val)
			}
			value.SetInt(i)
		case bool:
			b, err := strconv.ParseBool(val)
			if err != nil {
				log.Fatalf("Cannot convert %s to a boolean", val)
			}
			value.SetBool(b)
		}
	}

	harbor.Defaults(&base)
	if base.Trace {
		data, _ := yaml.Marshal(base)
		log.Infof("Using configuration: \n%s\n", console.StripSecrets(string(data)))
	}
	return base
}

func mergeConfigs(base *types.PlatformConfig, paths []string) error {
	for _, path := range paths {
		logger.Debugf("Merging %s", path)
		cfg := types.PlatformConfig{
			Source: path,
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "Failed to read config file %s", path)
		}

		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return errors.Wrap(err, "Failed to parse YAML")
		}

		for node, vm := range cfg.Nodes {
			if baseNode, ok := base.Nodes[node]; ok {
				if err := mergo.Merge(&baseNode, vm); err != nil {
					return errors.Wrapf(err, "Failed to merge nodes %s", node)
				}
				base.Nodes[node] = baseNode
			}
		}

		if err := mergo.Merge(base, cfg); err != nil {
			return errors.Wrapf(err, "Failed to merge in %s", path)
		}

		for _, config := range cfg.ImportConfigs {
			fullPath := filepath.Dir(path) + "/" + config
			if err := mergeConfigs(base, []string{fullPath}); err != nil {
				return err
			}
		}
	}

	return nil
}

func GlobalPreRun(cmd *cobra.Command, args []string) {
	level, _ := cmd.Flags().GetCount("loglevel")
	switch {
	case level > 1:
		log.SetLevel(log.TraceLevel)
	case level > 0:
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

var Render = &cobra.Command{
	Use:   "render",
	Short: "Generate kubeconfig files",
	Run: func(cmd *cobra.Command, args []string) {
		base := getConfig(cmd)
		data, _ := yaml.Marshal(base)
		fmt.Println(string(data))
	},
}
