package cmd

import (
	"fmt"
	"bytes"
	"strings"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

/*	deploy_base "github.com/moshloop/platform-cli/pkg/phases/base"
	"github.com/moshloop/platform-cli/pkg/phases/calico"
	"github.com/moshloop/platform-cli/pkg/phases/dex"
	"github.com/moshloop/platform-cli/pkg/phases/flux"
	"github.com/moshloop/platform-cli/pkg/phases/harbor"
	"github.com/moshloop/platform-cli/pkg/phases/monitoring"
	"github.com/moshloop/platform-cli/pkg/phases/nsx"
	"github.com/moshloop/platform-cli/pkg/phases/opa"
	"github.com/moshloop/platform-cli/pkg/phases/pgo"
	"github.com/moshloop/platform-cli/pkg/phases/stubs"
	"github.com/moshloop/platform-cli/pkg/phases/velero"*/
)

func checkErr(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

var Patch = &cobra.Command{
	Use:   "patch",
	Short: "Patch the platform",
	PreRun: func(cmd *cobra.Command, args []string) {
    	log.Infof("Inside rootCmd PreRun with args: %v\n", args)
    	patchFilePath, _ := cmd.Flags().GetString("path")
		if(patchFilePath == "-"){
			log.Fatal("Path to patch files require")
		}
		_ , err := ioutil.ReadDir(patchFilePath)
	    checkErr(err)
    },
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun{
			log.Infof("Running a dry-run mode, no changes will be made")
		}

		patchFilePath, _ := cmd.Flags().GetString("path")
		files, _ := ioutil.ReadDir(patchFilePath)
		var buffer bytes.Buffer
	    for _, file := range files {
	    	if(strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")){
	    		buffer.WriteString("  - ")
	    		buffer.WriteString(file.Name())
	    		buffer.WriteString("\n")
	    	}
	    }
	    patchFiles := buffer.String()

	    kustTemplate, err := ioutil.ReadFile("overlays/overlays_kustomization_template")
    	checkErr(err)

    	kustomization := fmt.Sprintf(string(kustTemplate), patchFiles)

    	fmt.Println(kustomization)

    	err = ioutil.WriteFile(patchFilePath+"/kustomization.yaml", []byte(kustomization), 0644)
    	checkErr(err)

    	if dryRun{
    		log.Infof("Running a dry-run mode")
    		
    	}
    	else{
    		log.Infof("Running a dry-run mode")
    	}
	},
}

func init() {

	/*dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		base.DryRun = true
		log.Infof("Running a dry-run mode, no changes will be made")
	}*/

	Patch.PersistentFlags().StringP("path", "p", "-", "Path to patch files")
	Patch.MarkFlagRequired("path")
}
