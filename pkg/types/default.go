package types

func DefaultPlatformConfig() PlatformConfig {
	config := PlatformConfig{
		Ldap: &Ldap{
			GroupObjectClass: "group",
			GroupNameAttr:    "name",
		},
		OAuth2Proxy: &OAuth2Proxy{
			DockerImage: "quay.io/pusher/oauth2_proxy",
			Version:     "v5.0.0",
		},
	}
	return config
}
