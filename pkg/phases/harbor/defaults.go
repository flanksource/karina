package harbor

import (
	"fmt"
	"strings"

	"github.com/flanksource/karina/pkg/types"
)

var dbCluster = "harbor"
var dbNames = []string{"registry", "trivy", "notary_server", "notary_signer"}

const Namespace = "harbor"

func Defaults(p *types.PlatformConfig) {
	harbor := p.Harbor
	if harbor.IsDisabled() {
		return
	}
	if harbor.LogLevel == "" {
		harbor.LogLevel = "warn"
	}
	if harbor.AdminPassword == "" {
		harbor.AdminPassword = "Harbor12345"
	}
	if harbor.URL == "" {
		harbor.URL = "https://harbor." + p.Domain
	}

	if !strings.HasPrefix(harbor.Version, "v1") {
		// from v2 all images use the same version label
		harbor.RegistryVersion = harbor.Version
		harbor.ChartVersion = harbor.Version
	}

	if harbor.RegistryVersion == "" {
		harbor.RegistryVersion = "v2.7.1-patch-2819-2553-" + harbor.Version
	}

	if harbor.Replicas == 0 {
		harbor.Replicas = 1
	}
	if p.Ldap != nil {
		if p.Ldap.GroupNameAttr == "" {
			p.Ldap.GroupNameAttr = "name"
		}
		if p.Ldap.GroupObjectClass == "" {
			p.Ldap.GroupObjectClass = "group"
		}
		settings := harbor.Settings
		if settings == nil {
			settings = &types.HarborSettings{}
		}
		settings.LdapURL = p.Ldap.GetConnectionURL()
		settings.LdapBaseDN = p.Ldap.UserDN
		verify := false
		settings.LdapVerifyCert = &verify
		settings.AuthMode = "ldap_auth"
		if settings.LdapUID == "" {
			settings.LdapUID = "sAMAccountName"
		}
		settings.LdapSearchPassword = p.Ldap.Password
		settings.LdapSearchDN = p.Ldap.Username
		if settings.LdapGroupSearchFilter == "" {
			settings.LdapGroupSearchFilter = fmt.Sprintf("objectclass=%s", p.Ldap.GroupObjectClass)
		}
		if settings.LdapGroupAdminDN == "" {
			settings.LdapGroupAdminDN = p.Ldap.AdminGroup
		}
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
