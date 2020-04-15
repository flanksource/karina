package opa

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/flanksource/commons/files"

	"github.com/moshloop/platform-cli/pkg/platform"
)

func readFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func deploy(platform *platform.Platform, policiesPath string) error {
	policyFiles, err := ioutil.ReadDir(policiesPath)
	if err != nil {
		return err
	}

	for _, policyFile := range policyFiles {
		policy, err := readFile(policiesPath + "/" + policyFile.Name())
		if err != nil {
			return fmt.Errorf("cannot read %s", policyFile)
		}
		if err := platform.CreateOrUpdateConfigMap(files.GetBaseName(policyFile.Name()), Namespace, map[string]string{
			policyFile.Name(): policy,
		}); err != nil {
			return fmt.Errorf("deploy: failed to create/update configmap: %v", err)
		}
	}
	return err
}
