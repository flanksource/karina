package types

func DefaultPlatformConfig() PlatformConfig {
	config := PlatformConfig{
		CertManager: CertManager{
			Version: "v1.6.1",
		},
		TemplateOperator: TemplateOperator{
			XDisabled: XDisabled{Version: "v0.1.9"},
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
		Nginx: &Nginx{
			Default: true,
			Version: "v1.1.1",
		},
	}
	return config
}
