package harbor

import (
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
)

var dbCluster = "harbor"
var dbNames = []string{"registry", "clair", "notary_server", "notary_signer"}

func defaults(p *platform.Platform) {
	harbor := p.Harbor
	if harbor.AdminPassword == "" {
		harbor.AdminPassword = "Harbor12345"
	}
	if harbor.URL == "" {
		harbor.URL = "https://harbor." + p.Domain
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
		harbor.ChartVersion = "v1.2.0"
	}
	p.Harbor = harbor
}
