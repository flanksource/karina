package certmanager

import (
	"time"
)

// IssuerConditionType represents an Issuer condition value.
type IssuerConditionType string

const (
	// IssuerConditionReady represents the fact that a given Issuer condition
	// is in ready state.
	IssuerConditionReady IssuerConditionType = "Ready"
)

// CertificateConditionType represents an Certificate condition value.
type CertificateConditionType string

const (
	// CertificateConditionReady indicates that a certificate is ready for use.
	// This is defined as:
	// - The target secret exists
	// - The target secret contains a certificate that has not expired
	// - The target secret contains a private key valid for the certificate
	// - The commonName and dnsNames attributes match those specified on the Certificate
	CertificateConditionReady CertificateConditionType = "Ready"
)

const (
	AltNamesAnnotationKey    = "cert-manager.io/alt-names"
	IPSANAnnotationKey       = "cert-manager.io/ip-sans"
	URISANAnnotationKey      = "cert-manager.io/uri-sans"
	CommonNameAnnotationKey  = "cert-manager.io/common-name"
	IssuerNameAnnotationKey  = "cert-manager.io/issuer-name"
	IssuerKindAnnotationKey  = "cert-manager.io/issuer-kind"
	IssuerGroupAnnotationKey = "cert-manager.io/issuer-group"
	CertificateNameKey       = "cert-manager.io/certificate-name"
)

// Deprecated annotation names for Secrets
const (
	DeprecatedIssuerNameAnnotationKey = "certmanager.k8s.io/issuer-name"
	DeprecatedIssuerKindAnnotationKey = "certmanager.k8s.io/issuer-kind"
)

const (
	// issuerNameAnnotation can be used to override the issuer specified on the
	// created Certificate resource.
	IngressIssuerNameAnnotationKey = "cert-manager.io/issuer"
	// clusterIssuerNameAnnotation can be used to override the issuer specified on the
	// created Certificate resource. The Certificate will reference the
	// specified *ClusterIssuer* instead of normal issuer.
	IngressClusterIssuerNameAnnotationKey = "cert-manager.io/cluster-issuer"
	// acmeIssuerHTTP01IngressClassAnnotation can be used to override the http01 ingressClass
	// if the challenge type is set to http01
	IngressACMEIssuerHTTP01IngressClassAnnotationKey = "acme.cert-manager.io/http01-ingress-class"

	// IngressClassAnnotationKey picks a specific "class" for the Ingress. The
	// controller only processes Ingresses with this annotation either unset, or
	// set to either the configured value or the empty string.
	IngressClassAnnotationKey = "kubernetes.io/ingress.class"
)

// Annotation names for CertificateRequests
const (
	CRPrivateKeyAnnotationKey = "cert-manager.io/private-key-secret-name"
)

const (
	// IssueTemporaryCertificateAnnotation is an annotation that can be added to
	// Certificate resources.
	// If it is present, a temporary internally signed certificate will be
	// stored in the target Secret resource whilst the real Issuer is processing
	// the certificate request.
	IssueTemporaryCertificateAnnotation = "cert-manager.io/issue-temporary-certificate"
)

const (
	ClusterIssuerKind      = "ClusterIssuer"
	IssuerKind             = "Issuer"
	CertificateKind        = "Certificate"
	CertificateRequestKind = "CertificateRequest"
)

const (
	// WantInjectAnnotation is the annotation that specifies that a particular
	// object wants injection of CAs.  It takes the form of a reference to a certificate
	// as namespace/name.  The certificate is expected to have the is-serving-for annotations.
	WantInjectAnnotation = "cert-manager.io/inject-ca-from"

	// WantInjectAPIServerCAAnnotation, if set to "true", will make the cainjector
	// inject the CA certificate for the Kubernetes apiserver into the resource.
	// It discovers the apiserver's CA by inspecting the service account credentials
	// mounted into the cainjector pod.
	WantInjectAPIServerCAAnnotation = "cert-manager.io/inject-apiserver-ca"

	// WantInjectFromSecretAnnotation is the annotation that specifies that a particular
	// object wants injection of CAs.  It takes the form of a reference to a Secret
	// as namespace/name.
	WantInjectFromSecretAnnotation = "cert-manager.io/inject-ca-from-secret"

	// AllowsInjectionFromSecretAnnotation is an annotation that must be added
	// to Secret resource that want to denote that they can be directly
	// injected into injectables that have a `inject-ca-from-secret` annotation.
	// If an injectable references a Secret that does NOT have this annotation,
	// the cainjector will refuse to inject the secret.
	AllowsInjectionFromSecretAnnotation = "cert-manager.io/allow-direct-injection"
)

// Issuer specific Annotations
const (
	// VenafiCustomFieldsAnnotationKey is the annotation that passes on JSON encoded custom fields to the Venafi issuer
	// This will only work with Venafi TPP v19.3 and higher
	// The value is an array with objects containing the name and value keys
	// for example: `[{"name": "custom-field", "value": "custom-value"}]`
	VenafiCustomFieldsAnnotationKey = "venafi.cert-manager.io/custom-fields"
)

// KeyUsage specifies valid usage contexts for keys.
// See: https://tools.ietf.org/html/rfc5280#section-4.2.1.3
//      https://tools.ietf.org/html/rfc5280#section-4.2.1.12
// Valid KeyUsage values are as follows:
// "signing",
// "digital signature",
// "content commitment",
// "key encipherment",
// "key agreement",
// "data encipherment",
// "cert sign",
// "crl sign",
// "encipher only",
// "decipher only",
// "any",
// "server auth",
// "client auth",
// "code signing",
// "email protection",
// "s/mime",
// "ipsec end system",
// "ipsec tunnel",
// "ipsec user",
// "timestamping",
// "ocsp signing",
// "microsoft sgc",
// "netscape sgc"
// +kubebuilder:validation:Enum="signing";"digital signature";"content commitment";"key encipherment";"key agreement";"data encipherment";"cert sign";"crl sign";"encipher only";"decipher only";"any";"server auth";"client auth";"code signing";"email protection";"s/mime";"ipsec end system";"ipsec tunnel";"ipsec user";"timestamping";"ocsp signing";"microsoft sgc";"netscape sgc"
type KeyUsage string

const (
	UsageSigning            KeyUsage = "signing"
	UsageDigitalSignature   KeyUsage = "digital signature"
	UsageContentCommittment KeyUsage = "content commitment"
	UsageKeyEncipherment    KeyUsage = "key encipherment"
	UsageKeyAgreement       KeyUsage = "key agreement"
	UsageDataEncipherment   KeyUsage = "data encipherment"
	UsageCertSign           KeyUsage = "cert sign"
	UsageCRLSign            KeyUsage = "crl sign"
	UsageEncipherOnly       KeyUsage = "encipher only"
	UsageDecipherOnly       KeyUsage = "decipher only"
	UsageAny                KeyUsage = "any"
	UsageServerAuth         KeyUsage = "server auth"
	UsageClientAuth         KeyUsage = "client auth"
	UsageCodeSigning        KeyUsage = "code signing"
	UsageEmailProtection    KeyUsage = "email protection"
	UsageSMIME              KeyUsage = "s/mime"
	UsageIPsecEndSystem     KeyUsage = "ipsec end system"
	UsageIPsecTunnel        KeyUsage = "ipsec tunnel"
	UsageIPsecUser          KeyUsage = "ipsec user"
	UsageTimestamping       KeyUsage = "timestamping"
	UsageOCSPSigning        KeyUsage = "ocsp signing"
	UsageMicrosoftSGC       KeyUsage = "microsoft sgc"
	UsageNetscapeSGC        KeyUsage = "netscape sgc"
)

// DefaultKeyUsages contains the default list of key usages
func DefaultKeyUsages() []KeyUsage {
	// The serverAuth EKU is required as of Mac OS Catalina: https://support.apple.com/en-us/HT210176
	// Without this usage, certificates will _always_ flag a warning in newer Mac OS browsers.
	return []KeyUsage{UsageDigitalSignature, UsageKeyEncipherment, UsageServerAuth}
}

const (
	// minimum permitted certificate duration by cert-manager
	MinimumCertificateDuration = time.Hour

	// default certificate duration if Issuer.spec.duration is not set
	DefaultCertificateDuration = time.Hour * 24 * 90

	// minimum certificate duration before certificate expiration
	MinimumRenewBefore = time.Minute * 5

	// Default duration before certificate expiration if  Issuer.spec.renewBefore is not set
	DefaultRenewBefore = time.Hour * 24 * 30
)

const (
	// Default index key for the Secret reference for Token authentication
	DefaultVaultTokenAuthSecretKey = "token"

	// Default mount path location for Kubernetes ServiceAccount authentication
	// (/v1/auth/kubernetes). The endpoint will then be called at `/login`, so
	// left as the default, `/v1/auth/kubernetes/login` will be called.
	DefaultVaultKubernetesAuthMountPath = "/v1/auth/kubernetes"
)

// +kubebuilder:validation:Enum=rsa;ecdsa
type KeyAlgorithm string

const (
	RSAKeyAlgorithm   KeyAlgorithm = "rsa"
	ECDSAKeyAlgorithm KeyAlgorithm = "ecdsa"
)

// +kubebuilder:validation:Enum=pkcs1;pkcs8
type KeyEncoding string

const (
	PKCS1 KeyEncoding = "pkcs1"
	PKCS8 KeyEncoding = "pkcs8"
)
