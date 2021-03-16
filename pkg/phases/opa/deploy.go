package opa

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
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

func deployTemplates(p *platform.Platform, path string) error {
	return errors.Wrap(deployManifests(p, path), "failed to deploy templates")
}

func deployConstraints(p *platform.Platform, path string) error {
	return errors.Wrap(deployManifests(p, path), "failed to deploy constraints")
}

func deployManifests(p *platform.Platform, path string) error {
	if path == "" {
		return nil
	}
	manifests, err := ioutil.ReadDir(path)
	if err != nil {
		return errors.Wrapf(err, "failed to read directory: %s", path)
	}

	for _, manifestFile := range manifests {
		manifest, err := readFile(path + "/" + manifestFile.Name())
		if err != nil {
			return errors.Wrapf(err, "failed to read file %s", manifestFile.Name())
		}
		items, err := kommons.GetUnstructuredObjects([]byte(manifest))
		if err != nil {
			return err
		}
		for _, item := range items {
			if strings.HasPrefix(item.GetKind(), "constraints.gatekeeper.sh") {
				// wait for the Gatekeeper webhook to be ready
				if _, err := p.WaitForResource("ConstraintTemplate", v1.NamespaceAll, strings.ToLower(item.GetName()), 2*time.Minute); err != nil {
					return err
				}
			}
		}
		if err := p.ApplyUnstructured(v1.NamespaceAll, items...); err != nil {
			return errors.Wrapf(err, "failed to apply file %s", manifestFile.Name())
		}
	}
	return nil
}
