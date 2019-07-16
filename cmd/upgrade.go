package cmd

import (
	"github.com/spf13/cobra"
)

var Upgrade = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade the core platform components to their latest versions",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		//kubeadm config print init-defaults
		// kubeadm config print join-defaults
		//https://docs.projectcalico.org/v3.8/manifests/calico.yaml
		// https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/static/mandatory.yaml
		//https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/static/provider/baremetal/service-nodeport.yaml
		//https://raw.githubusercontent.com/kubernetes/dashboard/v1.10.1/src/deploy/recommended/kubernetes-dashboard.yaml
		// https: //raw.githubusercontent.com/heptiolabs/eventrouter/master/yaml/eventrouter.yaml
	},
}
