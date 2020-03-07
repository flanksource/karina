package types

func DefaultPlatformConfig() PlatformConfig {
	config := PlatformConfig{
		Ldap: &Ldap{
			GroupObjectClass: "group",
			GroupNameAttr:    "name",
		},
	}
	return config
}
