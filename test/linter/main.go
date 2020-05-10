package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"gopkg.in/flanksource/yaml.v3"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	configFile = "test/linter.yaml"
)

var (
	config = &Config{}
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
	log.Println("All checks passed!")
}
