package cmd

import (
	log "github.com/sirupsen/logrus"

	"github.com/moshloop/platform-cli/pkg/provision"
	"github.com/spf13/cobra"
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

		platform := getPlatform(cmd)
		//copy master config
		vm := platform.Master
		vm.MemoryGB = int64(mem)
		vm.DiskGB = disk
		vm.CPUs = int32(cpu)
		vm.Name = platform.HostPrefix + name

		if err := provision.VM(getPlatform(cmd), &vm, specs...); err != nil {
			log.Fatalf("Failed to provision vm, %s", err)
		}
	},
}

func init() {
	Provision.AddCommand(cluster, vm)
	vm.Flags().String("name", "", "Name of vm")
	vm.Flags().String("template", "", "template to use")
	vm.Flags().Int("mem", 8, "Memory in GB")
	vm.Flags().Int("cpu", 2, "Number of cpus")
	vm.Flags().Int("disk", 50, "Disk size in GB")
	vm.Flags().StringArray("konfig", nil, "One or more konfigadm specs")
}
