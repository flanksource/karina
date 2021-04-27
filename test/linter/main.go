package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/flanksource/kommons"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/flanksource/yaml.v3"
)

const (
	configFile = "test/linter.yaml"
)

var allowedDuplicateKeys = []string{"CustomResourceDefinition-servicemonitors.monitoring.coreos.com"}
var ignoreManifestsSubPaths = []string{
	"manifests/gatekeeper/(.*)",
	"manifests/harbor/(.*)",
	"manifests/monitoring/(.*)",
	"manifests/nginx/(.*)",
	"manifests/template/(.*)",
	"manifests/test/(.*)",
	"manifests/upstream/(.*)",
}
var (
	config = &Config{}
	keys   = make([]string, 10, 500)
)

type Config struct {
	YamlCheck YamlCheck `yaml:"yamlCheck,omitempty"`
}

type YamlCheck struct {
	Ignore []string `yaml:"ignore,omitempty"`
}

func checkPath(filePath string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if path.Ext(filePath) == ".yml" {
		// Check if the path is ignored
		for _, ignore := range config.YamlCheck.Ignore {
			re, err := regexp.Compile(fmt.Sprintf("^%s$", ignore))
			if err != nil {
				return errors.Wrapf(err, "failed to compile regex %s", ignore)
			}
			if re.Match([]byte(filePath)) {
				return nil
			}
		}
		return errors.Errorf("File %s should have yaml extension instead of yml", filePath)
	}
	return nil
}

func checkManifests(filePath string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if path.Ext(filePath) == ".yaml" {
		// Check if the path is ignored
		for _, ignore := range ignoreManifestsSubPaths {
			re, err := regexp.Compile(fmt.Sprintf("^%s$", ignore))
			if err != nil {
				return errors.Wrapf(err, "failed to compile regex %s", ignore)
			}
			if re.Match([]byte(filePath)) {
				return nil
			}
		}
		err := generateUniqueKeys(filePath)
		if err != nil {
			return err
		}
	}
	return nil
}

// Generates keys in form of kind-name-namespace or kind-name returns error in case key already exists
func generateUniqueKeys(manifest string) error {
	yamlFile, err := ioutil.ReadFile(manifest)
	if err != nil {
		return errors.Errorf("error reading the file %v", manifest)
	}
	manifestData, err := kommons.GetUnstructuredObjects(yamlFile)
	if err != nil {
		log.Warnf("error parsing the yaml %v", manifest)
	}
	for i := range manifestData {
		if manifestData[i].Object["metadata"].(map[string]interface{})["namespace"] != nil {
			value := fmt.Sprintf("%v-%v-%v", manifestData[i].Object["kind"].(string), manifestData[i].Object["metadata"].(map[string]interface{})["name"].(string), manifestData[i].Object["metadata"].(map[string]interface{})["namespace"].(string))
			if !contains(allowedDuplicateKeys, value) && contains(keys, value) {
				return errors.Errorf("error %v already present in the manifests", value)
			}
			keys = append(keys, value)
		} else {
			value := fmt.Sprintf("%v-%v", manifestData[i].Object["kind"].(string), manifestData[i].Object["metadata"].(map[string]interface{})["name"].(string))
			if !contains(allowedDuplicateKeys, value) && contains(keys, value) {
				return errors.Errorf("error %v already present in the manifests", value)
			}
			keys = append(keys, value)
		}
	}
	return nil
}

func main() {
	yamlBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "Failed to read %s", configFile))
	}
	if err := yaml.Unmarshal(yamlBytes, config); err != nil {
		log.Fatal(errors.Wrapf(err, "Failed to parse %s", configFile))
	}

	err = filepath.Walk(".", checkPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to list files"))
	}
	err = filepath.Walk("manifests/", checkManifests)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Manifest linting failed"))
	}
	log.Println("All checks passed!")
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
