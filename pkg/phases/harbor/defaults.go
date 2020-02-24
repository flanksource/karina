package harbor

import (
	"strings"

	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
)

var dbCluster = "harbor"
var dbNames = []string{"registry", "clair", "notary_server", "notary_signer"}

var clairVersions = map[string]string{
	"latest": "2.1.1",
	"v1.9.0": "2.0.9",
	"v1.9.1": "2.0.9",
	"v1.9.2": "2.0.9",
	"v1.9.3": "2.1.0",
	"v1.9.4": "2.1.0",
	"v1.8.5": "2.0.8",
	"v1.8.4": "2.0.8",
	"v1.8.3": "2.0.8",
	"v1.8.2": "2.0.8",
	"v1.8.6": "2.1.0",
}

const Namespace = "harbor"

func defaults(p *platform.Platform) {
	harbor := p.Harbor
	if harbor == nil {
		return
	}
	if harbor.AdminPassword == "" {
		harbor.AdminPassword = "Harbor12345"
	}
	if harbor.URL == "" {
		harbor.URL = "https://harbor." + p.Domain
	}

	if harbor.ClairVersion == "" {
		if version, ok := clairVersions[harbor.Version]; ok {
			harbor.ClairVersion = version
		} else {
			harbor.ClairVersion = clairVersions["latest"]
		}
	}
	if harbor.RegistryVersion == "" {
	 	harbor.RegistryVersion = "v2.7.1-patch-2819-2553-"+ harbor.Version
	}

	if harbor.Replicas == 0 {
		harbor.Replicas = 1
	}
	if p.Ldap != nil {
		settings := harbor.Settings
		if settings == nil {
			settings = &types.HarborSettings{}
		}
		settings.LdapURL = "ldaps://" + p.Ldap.Host
		settings.LdapBaseDN = p.Ldap.BindDN
		verify := false
		settings.LdapVerifyCert = &verify
		settings.AuthMode = "ldap_auth"
		settings.LdapUID = "sAMAccountName"
		settings.LdapSearchPassword = p.Ldap.Password
		settings.LdapSearchDN = p.Ldap.Username
		settings.LdapGroupSearchFilter = "objectclass=group"
		settings.LdapGroupAdminDN = p.Ldap.AdminGroup
		harbor.Settings = settings
	}
	if harbor.ChartVersion == "" {
		if strings.HasPrefix(harbor.Version, "v1.10") {
			harbor.ChartVersion = "v1.3.0"
		} else if strings.HasPrefix(harbor.Version, "v1.9") {
			harbor.ChartVersion = "v1.2.0"
		} else if strings.HasPrefix(harbor.Version, "v1.8") {
			harbor.ChartVersion = "v1.1.0"
		} else {
			harbor.ChartVersion = "v1.3.0"
		}
	}
	p.Harbor = harbor
}
