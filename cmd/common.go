package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/files"

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
	sops "go.mozilla.org/sops/v3/decrypt"
	yaml "gopkg.in/flanksource/yaml.v3"
)

type configMerger struct {
	ReadFunction func(path string) ([]byte, error)
}

func (c configMerger) MergeConfigs(base *types.PlatformConfig, paths []string) error {
	startWD, err := os.Getwd()
	if err != nil {
		return err
	}
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		dir := filepath.Dir(absPath)
		err = os.Chdir(dir)
		if err != nil {
			return err
		}
		data, err := c.ReadFunction(absPath)
		if err != nil {
			return errors.Wrapf(err, "Failed to read config file %s", absPath)
		}
		err = mergeConfigBytes(base, data, absPath)
		if err != nil {
			return errors.Wrapf(err, "Failed to merge config file %s", absPath)
		}
		err = os.Chdir(startWD)
		if err != nil {
			return err
		}
	}
	return nil
}

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
	inCluster, _ := cmd.Flags().GetBool("in-cluster")
	skipDecryption, _ := cmd.Flags().GetBool("skip-decrypt")
	prune, _ := cmd.Flags().GetBool("prune")

	base := types.PlatformConfig{
		E2E:             e2e,
		Trace:           trace,
		DryRun:          dryRun,
		SkipDecrypt:     skipDecryption,
		InClusterConfig: inCluster,
		Prune:           prune,
	}
	return NewConfigFromBase(base, paths, extras)
}

// Deprecated: use  NewConfigFromBase instead
func NewConfig(paths []string, extras []string) types.PlatformConfig {
	return NewConfigFromBase(types.PlatformConfig{}, paths, extras)
}

func NewConfigFromBase(base types.PlatformConfig, paths []string, extras []string) types.PlatformConfig {
	splitPaths := []string{}
	for _, path := range paths {
		splitPaths = append(splitPaths, strings.Split(path, ",")...)
	}

	if len(paths) == 0 {
		log.Fatalf("Must specify at least 1 config")
	}
	paths = splitPaths
	base.Source = paths[0]
	cm := configMerger{
		ReadFunction: func(path string) ([]byte, error) { return ioutil.ReadFile(path) },
	}
	if err := cm.MergeConfigs(&base, paths); err != nil {
		log.Fatalf(err.Error())
	}
	if err := os.Chdir(filepath.Dir(base.Source)); err != nil {
		log.Fatalf("Could not enter config folder")
	}

	defaultConfig := types.DefaultPlatformConfig()
	if defaultConfig.Kubernetes.Managed {
		defaultConfig.Dex.Disabled = true
		defaultConfig.LocalPath.Disabled = true
		defaultConfig.Calico.Disabled = true
	}
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
		if err := lookup.Set(&base, key, val); err != nil {
			log.Fatalf("failed to set key %s: %v", key, err)
		}
	}

	harbor.Defaults(&base)
	if base.Trace {
		data, _ := yaml.Marshal(base)
		log.Infof("Using configuration: \n%s\n", console.StripSecrets(string(data)))
	}
	return base
}

func mergeConfigBytes(base *types.PlatformConfig, data []byte, path string) error {
	logger.Debugf("Merging %s", path)
	cfg := types.PlatformConfig{
		Source: path,
	}

	reader := bytes.NewReader(data)
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true)

	if err := decoder.Decode(&cfg); err != nil {
		return errors.Wrap(err, "Failed to parse YAML")
	}

	for k := range cfg.Patches {
		if files.IsValidPathType(cfg.Patches[k], "yaml", "yml", "json") {
			absPath, err := filepath.Abs(cfg.Patches[k])
			if err != nil {
				return err
			}
			cfg.Patches[k] = absPath
		}
	}
	pathList := []*string{
		&cfg.TrustedCA,
		&cfg.Gatekeeper.Templates, &cfg.Gatekeeper.Constraints,
		&cfg.Kubernetes.AuditConfig.PolicyFile, &cfg.Kubernetes.EncryptionConfig.EncryptionProviderConfigFile,
	}
	if cfg.IngressCA != nil {
		pathList = append(pathList, &cfg.IngressCA.Cert, &cfg.IngressCA.PrivateKey)
	}
	if cfg.CA != nil {
		pathList = append(pathList, &cfg.CA.Cert, &cfg.CA.PrivateKey)
	}
	if cfg.SealedSecrets.Certificate != nil {
		pathList = append(pathList, &cfg.SealedSecrets.Certificate.Cert, &cfg.SealedSecrets.Certificate.PrivateKey)
	}

	for _, field := range pathList {
		if err := MakeAbsolute(field); err != nil {
			return err
		}
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

	tcm := configMerger{
		ReadFunction: func(path string) ([]byte, error) { return ioutil.ReadFile(path) },
	}
	scm := configMerger{
		ReadFunction: func(path string) ([]byte, error) {
			data, err := sops.File(path, "yaml")
			if err != nil {
				return nil, errors.WithMessage(err, "SOPS decryption failed, to skip use --skip-decrypt")
			}
			return data, nil
		},
	}

	for _, config := range cfg.ImportConfigs {
		fullPath := filepath.Dir(path) + "/" + config
		if err := tcm.MergeConfigs(base, []string{fullPath}); err != nil {
			return err
		}
	}

	for _, config := range cfg.ConfigFrom {
		if config.FilePath != "" {
			fullPath := filepath.Dir(path) + "/" + config.FilePath
			if err := tcm.MergeConfigs(base, []string{fullPath}); err != nil {
				return err
			}
		}

		if config.SopsPath != "" && !base.SkipDecrypt {
			fullPath := filepath.Dir(path) + "/" + config.SopsPath
			if err := scm.MergeConfigs(base, []string{fullPath}); err != nil {
				return err
			}
		} else if config.SopsPath != "" && base.SkipDecrypt {
			logger.Infof("Skipping decryption of sops file: %s", config.SopsPath)
		}
	}

	return nil
}

func MakeAbsolute(path *string) error {
	if path != nil {
		if *path != "" && is.File(*path) {
			absPath, err := filepath.Abs(*path)
			if err != nil {
				return err
			}
			*path = absPath
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
