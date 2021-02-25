package types

func DefaultPlatformConfig() PlatformConfig {
	config := PlatformConfig{
		CertManager: CertManager{
			Version: "v1.0.3",
		},
		TemplateOperator: TemplateOperator{
			Disabled: Disabled{Version: "v0.1.9"},
		},
		Ldap: &Ldap{
			GroupObjectClass: "group",
			GroupNameAttr:    "name",
		},
		Kubernetes: Kubernetes{
			APIServerExtraArgs:  map[string]string{},
			ControllerExtraArgs: map[string]string{},
			SchedulerExtraArgs:  map[string]string{},
			KubeletExtraArgs:    map[string]string{},
			EtcdExtraArgs:       map[string]string{},
			ContainerRuntime:    "docker",
			Managed:             false,
		},
	}
	return config
}
