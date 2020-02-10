package cmd

import (
	"os"
	"fmt"
	"bytes"
	"strings"
	"io/ioutil"
	"os/exec"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/moshloop/platform-cli/pkg/phases/patch"
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

		kustWorkingDir := "overlays/patchDir/"
		os.RemoveAll(kustWorkingDir)
		err := os.MkdirAll(kustWorkingDir, 0755)
    	checkErr(err)
		patchFilePath, _ := cmd.Flags().GetString("path")
		files, _ := ioutil.ReadDir(patchFilePath)
		var buffer bytes.Buffer
	    for _, file := range files {
	    	if(strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")){
	    		buffer.WriteString("  - ")
	    		buffer.WriteString(file.Name())
	    		buffer.WriteString("\n")
	    		fmt.Println(file.Name())
	    		input, err := ioutil.ReadFile(patchFilePath+"/"+file.Name())
				checkErr(err)
				err = ioutil.WriteFile(kustWorkingDir+file.Name(), input, 0644)
				checkErr(err)
	    	}
	    }
	    patchFiles := buffer.String()

	    kustTemplate, err := ioutil.ReadFile("overlays/overlays_kustomization_template")
    	checkErr(err)

    	kustomizationYaml := fmt.Sprintf(string(kustTemplate), patchFiles)

    	err = ioutil.WriteFile(kustWorkingDir+"/kustomization.yaml", []byte(kustomizationYaml), 0644)
    	checkErr(err)

    	output, err := exec.Command("kustomize", "build", kustWorkingDir).Output()
    	fmt.Println("After runnnnnnnnn================")
    	checkErr(err)
    	finalPatchYaml := string(output[:])

    	if dryRun{
    		log.Infof("Yaml to apply")
    		fmt.Println(finalPatchYaml)
    	} else{
    		err = ioutil.WriteFile(kustWorkingDir+"/finalPatch.yaml", []byte(finalPatchYaml), 0644)
			checkErr(err)
    		log.Infof("Patching configs")
    		if err := patch.Install(getPlatform(cmd), kustWorkingDir+"/finalPatch.yaml"); err != nil {
				log.Fatalf("Error in patching: %s\n", err)
			}
    	}
	},
}

func init() {

	Patch.PersistentFlags().StringP("path", "p", "-", "Path to patch files")
	Patch.MarkFlagRequired("path")
}
