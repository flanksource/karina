package opa

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/flanksource/commons/files"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/pkg/errors"
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

func deployTemplates(p *platform.Platform, path string) error {
	return errors.Wrap(deployManifests(p, path), "failed to deploy templates")
}

func deployConstrains(p *platform.Platform, path string) error {
	return errors.Wrap(deployManifests(p, path), "failed to deploy constrains")
}

func deployManifests(p *platform.Platform, path string) error {
	manifests, err := ioutil.ReadDir(path)
	if err != nil {
		return errors.Wrapf(err, "failed to read directory: %s", path)
	}

	for _, manifestFile := range manifests {
		manifest, err := readFile(path + "/" + manifestFile.Name())
		if err != nil {
			return errors.Wrapf(err, "failed to read file %s", manifestFile.Name())
		}

		if err := p.ApplyText("", manifest); err != nil {
			return errors.Wrapf(err, "failed to apply file %s", manifestFile.Name())
		}
	}

	return nil
}
