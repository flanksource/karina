package ingress

import (
	"strings"

	"github.com/flanksource/commons/text"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
	log "github.com/sirupsen/logrus"
)

func IngressNginxAccessSnippet(platform *platform.Platform, c types.LdapAccessConfig) string {
	if !c.Enabled {
		var s = `
auth_request_set $authHeader0 $upstream_http_x_auth_request_user;
proxy_set_header 'x-auth-request-user' $authHeader0;
auth_request_set $authHeader1 $upstream_http_x_auth_request_email;
proxy_set_header 'x-auth-request-email' $authHeader1;
auth_request_set $authHeader2 $upstream_http_authorization;
proxy_set_header 'authorization' $authHeader2;
`
		return escapeString(s)
	}

	raw, err := platform.GetResourceByName("ldap_access.tmpl", "manifests")
	if err != nil {
		log.Fatalf("Failed to open ldap_access.tmpl: %v", err)
		return ""
	}

	groupsEscaped := make([]string, len(c.Groups))
	for i, group := range c.Groups {
		groupsEscaped[i] = "\"" + group + "\""
	}

	data := map[string]string{
		"groups": strings.Join(groupsEscaped, ","),
	}

	template, err := text.Template(raw, data)
	if err != nil {
		log.Fatalf("Failed to open ldap_access.tmpl: %v", err)
		return ""
	}

	return escapeString(template)
}

func escapeString(s string) string {
	withoutNewlines := strings.ReplaceAll(s, "\n", "\\n")
	withoutQuotes := strings.ReplaceAll(withoutNewlines, "\"", "\\\"")
	return withoutQuotes
}
