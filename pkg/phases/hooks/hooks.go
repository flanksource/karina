package hooks

import (
	"errors"
	"fmt"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func ApplyEnvVar(p *platform.Platform, envVar kommons.EnvVar) error {
	var namespace string
	if p.InClusterConfig {
		data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			return err
		}
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			namespace = ns
		} else {
			namespace = metav1.NamespaceAll
		}
	} else {
		namespace = "platform-system"
	}
	_, manifest, err := p.GetEnvValue(envVar, namespace)
	if err != nil {
		return err
	}
	return p.ApplyText("", manifest)
}

func ApplyHook(p *platform.Platform, name string, phase string) error {
	hook, ok := p.Hooks[name]
	if !ok {
		return nil
	}
	if phase == "before" {
		if (hook.Before == kommons.EnvVar{}) {
			return nil
		}
		return ApplyEnvVar(p, hook.Before)
	} else if phase == "after"{
		if (hook.After == kommons.EnvVar{}){
			return nil
		}
		return ApplyEnvVar(p, hook.After)
	} else {
		return errors.New(fmt.Sprintf("hook %v invalid.  Must be 'before' or 'after'", phase))
	}
	return nil
}

func ApplyBeforeHook(p *platform.Platform, name string) error {
	return ApplyHook(p, name, "before")
}

func ApplyAfterHook(p *platform.Platform, name string) error {
	return ApplyHook(p, name, "after")
}
