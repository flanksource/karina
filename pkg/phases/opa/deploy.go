package opa

import (
	log "github.com/sirupsen/logrus"
	"github.com/flanksource/commons/files"
	"io/ioutil"
	"os"
	"github.com/moshloop/platform-cli/pkg/platform"
)

func readFile(filename string ) string {
	file, err := os.Open(filename)
	if err != nil {
			log.Fatal(err)
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
			log.Fatal(err)
	}
	return string(data)
}

func Deploy(platform *platform.Platform, policiesPath string) error {
	
	policyFiles, err := ioutil.ReadDir(policiesPath)
	if err != nil {
			log.Fatal(err)
	}

	for _, policyFile := range policyFiles {
		if err := platform.CreateOrUpdateConfigMap(files.GetBaseName(policyFile.Name()), Namespace, map[string]string{
			policyFile.Name(): readFile(policiesPath+"/"+policyFile.Name()),
		}); err != nil {
			log.Tracef("Install: Failed to create/update configmap: %s", err)
			return err
		}	
	}
	return err
}

