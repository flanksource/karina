package cmd

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/moshloop/platform-cli/pkg/provision"
	"github.com/moshloop/platform-cli/pkg/provision/vmware"
)

var Provision = &cobra.Command{
	Use:   "provision",
	Short: "Commands for provisioning clusters and VMs",
}
var cluster = &cobra.Command{
	Use:   "cluster",
	Short: "Provision a new cluster",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := provision.Cluster(getPlatform(cmd)); err != nil {
			log.Fatalf("Failed to provision cluster, %s", err)
		}
	},
}

var vm = &cobra.Command{
	Use:   "vm",
	Short: "Provision a new vm",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		specs, _ := cmd.Flags().GetStringArray("konfig")
		name, _ := cmd.Flags().GetString("name")
		mem, _ := cmd.Flags().GetInt("mem")
		disk, _ := cmd.Flags().GetInt("disk")
		cpu, _ := cmd.Flags().GetInt("cpu")
		template, _ := cmd.Flags().GetString("template")
		network, _ := cmd.Flags().GetString("network")
		datastore, _ := cmd.Flags().GetString("datastore")
		pool, _ := cmd.Flags().GetString("resource-pool")
		dns, _ := cmd.Flags().GetString("dns")
		platform := getPlatform(cmd)
		//copy master config
		vm := platform.Master
		vmware.LoadGovcEnvVars(&vm)
		vm.MemoryGB = int64(mem)
		vm.DiskGB = disk
		vm.CPUs = int32(cpu)
		vm.Name = platform.HostPrefix + name
		if datastore != "" {
			vm.Datastore = datastore
		}
		if network != "" {
			vm.Network = strings.Split(network, ",")
		}
		if pool != "" {
			vm.ResourcePool = pool
		}
		if template != "" {
			vm.Template = template
		}

		if err := provision.VM(platform, &vm, specs...); err != nil {
			log.Fatalf("Failed to provision vm, %s", err)
		}
		if dns != "" {
			if err := platform.GetDNSClient().Append(dns, vm.IP); err != nil {
				log.Fatalf("Failed to update DNS %s => %s: %v", dns, vm.IP, err)
			}
		}

	},
}

func init() {
	Provision.AddCommand(cluster, vm)
	vm.Flags().String("name", "", "Name of vm")
	vm.Flags().String("dns", "", "DNS entry to add")
	vm.Flags().String("template", "", "template to use")
	vm.Flags().String("network", "", "network to use")
	vm.Flags().String("datastore", "", "datastore to")
	vm.Flags().String("resource-pool", "", "resource-pool")
	vm.Flags().Int("mem", 8, "Memory in GB")
	vm.Flags().Int("cpu", 2, "Number of cpus")
	vm.Flags().Int("disk", 50, "Disk size in GB")
	vm.Flags().StringArrayP("konfig", "k", []string{}, "One or more konfigadm specs")
}
