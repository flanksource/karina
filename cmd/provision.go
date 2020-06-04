package cmd

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/flanksource/karina/pkg/provision"
	"github.com/flanksource/karina/pkg/provision/vmware"
)

var Provision = &cobra.Command{
	Use:   "provision",
	Short: "Commands for provisioning clusters and VMs",
}
var vsphereCluster = &cobra.Command{
	Use:   "vsphere-cluster",
	Short: "Provision a new vsphere cluster",
	Args:  cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := getPlatform(cmd)
		if err := p.ValidateVSphereCluster(); err != nil {
			return fmt.Errorf("Failed to validate cluster configuration, %s", err)
		}
		if err := provision.VsphereCluster(p); err != nil {
			return fmt.Errorf("failed to provision cluster, %s", err)
		}
		return nil
	},
}

var kindCluster = &cobra.Command{
	Use:   "kind-cluster",
	Short: "Provision a new kind cluster",
	Args:  cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		p:= getPlatform(cmd)
		if err := p.ValidateKindCluster(); err != nil {
			return fmt.Errorf("Failed to validate cluster configuration, %s", err)
		}
		if err := provision.KindCluster(p); err != nil {
			return fmt.Errorf("failed to provision cluster, %s", err)
		}
		return nil
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
		vmware.LoadGovcEnvVars(*platform.Vsphere, &vm)
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
	Provision.AddCommand(vsphereCluster, kindCluster, vm)
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
