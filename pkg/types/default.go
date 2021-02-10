package types

func DefaultPlatformConfig() PlatformConfig {
	config := PlatformConfig{
		Calico: &Calico{
			IPIP:  "Never",
			VxLAN: "Never",
		},
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
			Managed:			 false,
		},
		EventRouter: EventRouter{
			FilebeatPrefix: "com.flanksource.infra",
		},
		PlatformOperator: &PlatformOperator{
			EnableClusterResourceQuota: false,
			WhitelistedPodAnnotations:  []string{"com.flanksource.infra.logs/enabled", "co.elastic.logs/enabled"},
		},
		Gatekeeper: Gatekeeper{
			AuditInterval: 60,
		},
	}
	return config
}
