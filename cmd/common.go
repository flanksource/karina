package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/flanksource/commons/console"
	"github.com/flanksource/commons/is"
	"github.com/flanksource/commons/lookup"
	"github.com/flanksource/commons/text"
	"github.com/flanksource/yaml"
	"github.com/imdario/mergo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
)

func getPlatform(cmd *cobra.Command) *platform.Platform {
	platform := platform.Platform{
		PlatformConfig: getConfig(cmd),
	}
	platform.Init()
	return &platform
}

func getConfig(cmd *cobra.Command) types.PlatformConfig {
	paths, _ := cmd.Flags().GetStringArray("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	extras, _ := cmd.Flags().GetStringArray("extra")
	trace, _ := cmd.Flags().GetBool("trace")

	return NewConfig(paths, dryRun, extras, showConfig, trace)
}

func NewConfig(paths []string, dryRun bool, extras []string, showConfig bool, trace bool) types.PlatformConfig {
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

	if err := mergeConfigs(&base, base.ImportConfigs); err != nil {
		log.Fatalf("Failed to merge configs: %v", err)
	}

	defaultConfig := types.DefaultPlatformConfig()
	if err := mergo.Merge(&base, defaultConfig); err != nil {
		log.Fatalf("Failed to merge default config, %v", err)
	}

	if dryRun {
		base.DryRun = true
		log.Infof("Running a dry-run mode, no changes will be made")
	}

	if trace {
		base.Trace = true
	}

	base.S3.AccessKey = template(base.S3.AccessKey)
	base.S3.SecretKey = template(base.S3.SecretKey)

	ldap := base.Ldap
	if ldap.Port == "" {
		ldap.Port = "636"
	}
	if ldap != nil {
		ldap.Username = template(ldap.Username)
		ldap.Password = template(ldap.Password)
		base.Ldap = ldap
	}

	base.Master.Network = templateSlice(base.Master.Network)
	base.Master.Cluster = template(base.Master.Cluster)
	base.Master.Template = template(base.Master.Template)

	nodes := base.Nodes
	for name, vm := range base.Nodes {
		vm.Network = templateSlice(vm.Network)
		vm.Cluster = template(vm.Cluster)
		vm.Template = template(vm.Template)
		nodes[name] = vm
	}
	base.Nodes = nodes

	dns := base.DNS
	if dns != nil {
		dns.Key = template(dns.Key)
		dns.KeyName = template(dns.KeyName)
		dns.AccessKey = template(dns.AccessKey)
		dns.SecretKey = template(dns.SecretKey)
		base.DNS = dns
	}

	if base.TrustedCA != "" && !is.File(base.TrustedCA) {
		base.TrustedCA = text.ToFile(base.TrustedCA, ".pem")
	}

	if base.NSX != nil && base.NSX.NsxV3 != nil {
		base.NSX.NsxV3.NsxAPIUser = template(base.NSX.NsxV3.NsxAPIUser)
		base.NSX.NsxV3.NsxAPIPass = template(base.NSX.NsxV3.NsxAPIPass)
	}

	if base.CA != nil {
		base.CA.Password = template(base.CA.Password)
	}
	if base.IngressCA != nil {
		base.IngressCA.Password = template(base.IngressCA.Password)
	}

	if base.FluentdOperator != nil {
		base.FluentdOperator.Elasticsearch.Password = template(base.FluentdOperator.Elasticsearch.Password)
	}

	if base.Filebeat != nil && base.Filebeat.Elasticsearch != nil {
		base.Filebeat.Elasticsearch.Password = template(base.Filebeat.Elasticsearch.Password)
	}

	if base.Vault != nil {
		base.Vault.AccessKey = template(base.Vault.AccessKey)
		base.Vault.SecretKey = template(base.Vault.SecretKey)
		base.Vault.KmsKeyID = template(base.Vault.KmsKeyID)
		base.Vault.Token = template(base.Vault.Token)
	}

	if base.CertManager.Vault != nil {
		base.CertManager.Vault.Token = template(base.CertManager.Vault.Token)
	}

	gitops := base.GitOps
	for i := range gitops {
		gitops[i].GitKey = template(gitops[i].GitKey)
	}
	base.GitOps = gitops
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

	if base.Trace {
		data, _ := yaml.Marshal(base)
		log.Infof("Using configuration: \n%s\n", console.StripSecrets(string(data)))
	}
	return base
}

func mergeConfigs(base *types.PlatformConfig, paths []string) error {
	for _, path := range paths {
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
	}

	return nil
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

func templateSlice(vals []string) []string {
	var out []string
	for _, val := range vals {
		if strings.HasPrefix(val, "$") {
			env := os.Getenv(val[1:])
			out = append(out, strings.Split(env, ",")...)
		} else {
			out = append(out, strings.Split(val, ",")...)
		}
	}
	return out
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
