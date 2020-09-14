package types

func DefaultPlatformConfig() PlatformConfig {
	config := PlatformConfig{
		Calico: &Calico{
			IPIP:  "Never",
			VxLAN: "Never",
		},
		CertManager: CertManager{
			Version: "v0.12.0",
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
		},
		EventRouter: EventRouter{
			FilebeatPrefix: "com.flanksource.infra",
		},
		PlatformOperator: &PlatformOperator{
			EnableClusterResourceQuota: false,
			WhitelistedPodAnnotations:  []string{"com.flanksource.infra.logs/enabled", "co.elastic.logs/enabled"},
		},
	}
	return config
}
