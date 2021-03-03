package opa

import (
	"io/ioutil"
	"os"

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

		if err := p.ApplyText("", manifest); err != nil {
			return errors.Wrapf(err, "failed to apply file %s", manifestFile.Name())
		}
	}

	return nil
}
