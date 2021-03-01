package certmanager

/*
Copyright 2019 The Jetstack cert-manager contributors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Certificate is a type to represent a Certificate from ACME
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Secret",type="string",JSONPath=".spec.secretName",description=""
// +kubebuilder:printcolumn:name="Issuer",type="string",JSONPath=".spec.issuerRef.name",description="",priority=1
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",priority=1
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=certificates,shortName=cert;certs
type Certificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertificateSpec   `json:"spec,omitempty"`
	Status CertificateStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertificateList is a list of Certificates
type CertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Certificate `json:"items"`
}

// CertificateSpec defines the desired state of Certificate.
// A valid Certificate requires at least one of a CommonName, DNSName, or
// URISAN to be valid.
type CertificateSpec struct {
	// Full X509 name specification (https://golang.org/pkg/crypto/x509/pkix/#Name).
	// +optional
	Subject *X509Subject `json:"subject,omitempty"`

	// CommonName is a common name to be used on the Certificate.
	// The CommonName should have a length of 64 characters or fewer to avoid
	// generating invalid CSRs.
	// This value is ignored by TLS clients when any subject alt name is set.
	// This is x509 behaviour: https://tools.ietf.org/html/rfc6125#section-6.4.4
	// +optional
	CommonName string `json:"commonName,omitempty"`

	// Certificate default Duration
	// +optional
	Duration *metav1.Duration `json:"duration,omitempty"`

	// Certificate renew before expiration duration
	// +optional
	RenewBefore *metav1.Duration `json:"renewBefore,omitempty"`

	// DNSNames is a list of subject alt names to be used on the Certificate.
	// +optional
	DNSNames []string `json:"dnsNames,omitempty"`

	// IPAddresses is a list of IP addresses to be used on the Certificate
	// +optional
	IPAddresses []string `json:"ipAddresses,omitempty"`

	// URISANs is a list of URI Subject Alternative Names to be set on this
	// Certificate.
	// +optional
	URISANs []string `json:"uriSANs,omitempty"`

	// EmailSANs is a list of Email Subject Alternative Names to be set on this
	// Certificate.
	// +optional
	EmailSANs []string `json:"emailSANs,omitempty"`

	// SecretName is the name of the secret resource to store this secret in
	SecretName string `json:"secretName"`

	// IssuerRef is a reference to the issuer for this certificate.
	// If the 'kind' field is not set, or set to 'Issuer', an Issuer resource
	// with the given name in the same namespace as the Certificate will be used.
	// If the 'kind' field is set to 'ClusterIssuer', a ClusterIssuer with the
	// provided name will be used.
	// The 'name' field in this stanza is required at all times.
	IssuerRef ObjectReference `json:"issuerRef"`

	// IsCA will mark this Certificate as valid for signing.
	// This implies that the 'cert sign' usage is set
	// +optional
	IsCA bool `json:"isCA,omitempty"`

	// Usages is the set of x509 actions that are enabled for a given key. Defaults are ('digital signature', 'key encipherment') if empty
	// +optional
	Usages []KeyUsage `json:"usages,omitempty"`

	// KeySize is the key bit size of the corresponding private key for this certificate.
	// If provided, value must be between 2048 and 8192 inclusive when KeyAlgorithm is
	// empty or is set to "rsa", and value must be one of (256, 384, 521) when
	// KeyAlgorithm is set to "ecdsa".
	// +kubebuilder:validation:ExclusiveMaximum=false
	// +kubebuilder:validation:Maximum=8192
	// +kubebuilder:validation:ExclusiveMinimum=false
	// +kubebuilder:validation:Minimum=0
	// +optional
	KeySize int `json:"keySize,omitempty"`

	// KeyAlgorithm is the private key algorithm of the corresponding private key
	// for this certificate. If provided, allowed values are either "rsa" or "ecdsa"
	// If KeyAlgorithm is specified and KeySize is not provided,
	// key size of 256 will be used for "ecdsa" key algorithm and
	// key size of 2048 will be used for "rsa" key algorithm.
	// +optional
	KeyAlgorithm KeyAlgorithm `json:"keyAlgorithm,omitempty"`

	// KeyEncoding is the private key cryptography standards (PKCS)
	// for this certificate's private key to be encoded in. If provided, allowed
	// values are "pkcs1" and "pkcs8" standing for PKCS#1 and PKCS#8, respectively.
	// If KeyEncoding is not specified, then PKCS#1 will be used by default.
	KeyEncoding KeyEncoding `json:"keyEncoding,omitempty"`
}

// X509Subject Full X509 name specification
type X509Subject struct {
	// Organizations to be used on the Certificate.
	// +optional
	Organizations []string `json:"organizations,omitempty"`
	// Countries to be used on the Certificate.
	// +optional
	Countries []string `json:"countries,omitempty"`
	// Organizational Units to be used on the Certificate.
	// +optional
	OrganizationalUnits []string `json:"organizationalUnits,omitempty"`
	// Cities to be used on the Certificate.
	// +optional
	Localities []string `json:"localities,omitempty"`
	// State/Provinces to be used on the Certificate.
	// +optional
	Provinces []string `json:"provinces,omitempty"`
	// Street addresses to be used on the Certificate.
	// +optional
	StreetAddresses []string `json:"streetAddresses,omitempty"`
	// Postal codes to be used on the Certificate.
	// +optional
	PostalCodes []string `json:"postalCodes,omitempty"`
	// Serial number to be used on the Certificate.
	// +optional
	SerialNumber string `json:"serialNumber,omitempty"`
}

// CertificateStatus defines the observed state of Certificate
type CertificateStatus struct {
	// +optional
	Conditions []CertificateCondition `json:"conditions,omitempty"`

	// +optional
	LastFailureTime *metav1.Time `json:"lastFailureTime,omitempty"`

	// The expiration time of the certificate stored in the secret named
	// by this resource in spec.secretName.
	// +optional
	NotAfter *metav1.Time `json:"notAfter,omitempty"`
}

// CertificateCondition contains condition information for an Certificate.
type CertificateCondition struct {
	// Type of the condition, currently ('Ready').
	Type CertificateConditionType `json:"type"`

	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	// +optional
	Message string `json:"message,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clusterissuers,scope=Cluster
type ClusterIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IssuerSpec   `json:"spec,omitempty"`
	Status IssuerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterIssuerList is a list of Issuers
type ClusterIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ClusterIssuer `json:"items"`
}

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=issuers
type Issuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IssuerSpec   `json:"spec,omitempty"`
	Status IssuerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IssuerList is a list of Issuers
type IssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Issuer `json:"items"`
}

// IssuerSpec is the specification of an Issuer. This includes any
// configuration required for the issuer.
type IssuerSpec struct {
	IssuerConfig `json:",inline"`
}

type IssuerConfig struct {

	// +optional
	CA *CAIssuer `json:"ca,omitempty"`

	// +optional
	Vault *VaultIssuer `json:"vault,omitempty"`

	// +optional
	SelfSigned *SelfSignedIssuer `json:"selfSigned,omitempty"`

	// +optional
	Venafi *VenafiIssuer `json:"venafi,omitempty"`

	// +optional
	Letsencrypt *LetsencryptIssuer `json:"acme,omitempty"`
}

// VenafiIssuer describes issuer configuration details for Venafi Cloud.
type VenafiIssuer struct {
	// Zone is the Venafi Policy Zone to use for this issuer.
	// All requests made to the Venafi platform will be restricted by the named
	// zone policy.
	// This field is required.
	Zone string `json:"zone"`

	// TPP specifies Trust Protection Platform configuration settings.
	// Only one of TPP or Cloud may be specified.
	// +optional
	TPP *VenafiTPP `json:"tpp,omitempty"`

	// Cloud specifies the Venafi cloud configuration settings.
	// Only one of TPP or Cloud may be specified.
	// +optional
	Cloud *VenafiCloud `json:"cloud,omitempty"`
}

// VenafiTPP defines connection configuration details for a Venafi TPP instance
type VenafiTPP struct {
	// URL is the base URL for the Venafi TPP instance
	URL string `json:"url"`

	// CredentialsRef is a reference to a Secret containing the username and
	// password for the TPP server.
	// The secret must contain two keys, 'username' and 'password'.
	CredentialsRef LocalObjectReference `json:"credentialsRef"`

	// CABundle is a PEM encoded TLS certificate to use to verify connections to
	// the TPP instance.
	// If specified, system roots will not be used and the issuing CA for the
	// TPP instance must be verifiable using the provided root.
	// If not specified, the connection will be verified using the cert-manager
	// system root certificates.
	// +optional
	CABundle []byte `json:"caBundle,omitempty"`
}

// VenafiCloud defines connection configuration details for Venafi Cloud
type VenafiCloud struct {
	// URL is the base URL for Venafi Cloud
	// +optional
	URL string `json:"url,omitempty"`

	// APITokenSecretRef is a secret key selector for the Venafi Cloud API token.
	APITokenSecretRef SecretKeySelector `json:"apiTokenSecretRef"`
}

// LetsencryptIssuer defines an issuer that uses the Letsencrypt API to issue
// certificates
type LetsencryptIssuer struct {
	// The API endpoint to use. Defaults to the production endpoint:
	// https://acme-v02.api.letsencrypt.org/directory
	Server        string            `json:"server"`
	Email         string            `json:"email"`
	PrivateKeyRef SecretKeySelector `json:"privateKeySecretRef,omitempty"`
	Solvers       []Solver          `json:"solvers"`
}

type Solver struct {
	DNS01  DNS01  `json:"dns01,omitempty"`
	HTTP01 HTTP01 `json:"http,omitempty"`
}

// The only DNS challenge Karina currently supports is Route53
type DNS01 struct {
	Route53 Route53 `json:"route53"`
}

type Route53 struct {
	Region             string            `json:"region"`
	HostedZoneID       string            `json:"hostedZoneID"`
	AccessKeyID        string            `json:"accessKeyID"`
	SecretAccessKeyRef SecretKeySelector `json:"secretAccessKeySecretRef"`
}

type HTTP01 struct {
	Type string `json:",inline"`
}

type SelfSignedIssuer struct{}

type VaultIssuer struct {
	// Vault authentication
	Auth VaultAuth `json:"auth"`

	// Server is the vault connection address
	Server string `json:"server"`

	// Vault URL path to the certificate role
	Path string `json:"path"`

	// Base64 encoded CA bundle to validate Vault server certificate. Only used
	// if the Server URL is using HTTPS protocol. This parameter is ignored for
	// plain HTTP protocol connection. If not set the system root certificates
	// are used to validate the TLS connection.
	// +optional
	CABundle []byte `json:"caBundle,omitempty"`
}

// Vault authentication  can be configured:
// - With a secret containing a token. Cert-manager is using this token as-is.
// - With a secret containing a AppRole. This AppRole is used to authenticate to
//   Vault and retrieve a token.
type VaultAuth struct {
	// This Secret contains the Vault token key
	// +optional
	TokenSecretRef *SecretKeySelector `json:"tokenSecretRef,omitempty"`

	// This Secret contains a AppRole and Secret
	// +optional
	AppRole *VaultAppRole `json:"appRole,omitempty"`

	// This contains a Role and Secret with a ServiceAccount token to
	// authenticate with vault.
	// +optional
	Kubernetes *VaultKubernetesAuth `json:"kubernetes,omitempty"`
}

type VaultAppRole struct {
	// Where the authentication path is mounted in Vault.
	Path string `json:"path"`

	// nolint: golint, stylecheck
	RoleId    string            `json:"roleId"`
	SecretRef SecretKeySelector `json:"secretRef"`
}

// Authenticate against Vault using a Kubernetes ServiceAccount token stored in
// a Secret.
type VaultKubernetesAuth struct {
	// The Vault mountPath here is the mount path to use when authenticating with
	// Vault. For example, setting a value to `/v1/auth/foo`, will use the path
	// `/v1/auth/foo/login` to authenticate with Vault. If unspecified, the
	// default value "/v1/auth/kubernetes" will be used.
	// +optional
	Path string `json:"mountPath,omitempty"`

	// The required Secret field containing a Kubernetes ServiceAccount JWT used
	// for authenticating with Vault. Use of 'ambient credentials' is not
	// supported.
	SecretRef SecretKeySelector `json:"secretRef"`

	// A required field containing the Vault Role to assume. A Role binds a
	// Kubernetes ServiceAccount with a set of Vault policies.
	Role string `json:"role"`
}

type CAIssuer struct {
	// SecretName is the name of the secret used to sign Certificates issued
	// by this Issuer.
	SecretName string `json:"secretName"`
}

// IssuerStatus contains status information about an Issuer
type IssuerStatus struct {
	// +optional
	Conditions []IssuerCondition `json:"conditions,omitempty"`
}

// IssuerCondition contains condition information for an Issuer.
type IssuerCondition struct {
	// Type of the condition, currently ('Ready').
	Type IssuerConditionType `json:"type"`

	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	// +optional
	Message string `json:"message,omitempty"`
}

type GenericIssuer interface {
	metav1.Object
	runtime.Object
	GetObjectMeta() *metav1.ObjectMeta
	GetSpec() *IssuerSpec
	GetStatus() *IssuerStatus
}

var _ GenericIssuer = &Issuer{}
var _ GenericIssuer = &ClusterIssuer{}

func (c *ClusterIssuer) GetObjectMeta() *metav1.ObjectMeta {
	return &c.ObjectMeta
}
func (c *ClusterIssuer) GetSpec() *IssuerSpec {
	return &c.Spec
}
func (c *ClusterIssuer) GetStatus() *IssuerStatus {
	return &c.Status
}
func (c *ClusterIssuer) SetSpec(spec IssuerSpec) {
	c.Spec = spec
}
func (c *ClusterIssuer) SetStatus(status IssuerStatus) {
	c.Status = status
}
func (c *ClusterIssuer) Copy() GenericIssuer {
	return c.DeepCopy()
}

func (c *Issuer) GetObjectMeta() *metav1.ObjectMeta {
	return &c.ObjectMeta
}
func (c *Issuer) GetSpec() *IssuerSpec {
	return &c.Spec
}
func (c *Issuer) GetStatus() *IssuerStatus {
	return &c.Status
}
func (c *Issuer) SetSpec(spec IssuerSpec) {
	c.Spec = spec
}
func (c *Issuer) SetStatus(status IssuerStatus) {
	c.Status = status
}
func (c *Issuer) Copy() GenericIssuer {
	return c.DeepCopy()
}

// ConditionStatus represents a condition's status.
// +kubebuilder:validation:Enum=True;False;Unknown
type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in
// the condition; "ConditionFalse" means a resource is not in the condition;
// "ConditionUnknown" means kubernetes can't decide if a resource is in the
// condition or not. In the future, we could add other intermediate
// conditions, e.g. ConditionDegraded.
const (
	// ConditionTrue represents the fact that a given condition is true
	ConditionTrue ConditionStatus = "True"

	// ConditionFalse represents the fact that a given condition is false
	ConditionFalse ConditionStatus = "False"

	// ConditionUnknown represents the fact that a given condition is unknown
	ConditionUnknown ConditionStatus = "Unknown"
)

type LocalObjectReference struct {
	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	// TODO: Add other useful fields. apiVersion, kind, uid?
	Name string `json:"name"`
}

// ObjectReference is a reference to an object with a given name, kind and group.
type ObjectReference struct {
	Name string `json:"name"`
	// +optional
	Kind string `json:"kind,omitempty"`
	// +optional
	Group string `json:"group,omitempty"`
}

type SecretKeySelector struct {
	// The name of the secret in the pod's namespace to select from.
	LocalObjectReference `json:",inline"`
	// The key of the secret to select from. Must be a valid secret key.
	// +optional
	Key string `json:"key,omitempty"`
}

const (
	TLSCAKey = "ca.crt"
)

const (
	CertificateRequestReasonPending = "Pending"
	CertificateRequestReasonFailed  = "Failed"
	CertificateRequestReasonIssued  = "Issued"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertificateRequest is a type to represent a Certificate Signing Request
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Issuer",type="string",JSONPath=".spec.issuerRef.name",description="",priority=1
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",priority=1
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=certificaterequests,shortName=cr;crs
type CertificateRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertificateRequestSpec   `json:"spec,omitempty"`
	Status CertificateRequestStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertificateRequestList is a list of Certificates
type CertificateRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CertificateRequest `json:"items"`
}

// CertificateRequestSpec defines the desired state of CertificateRequest
type CertificateRequestSpec struct {
	// Requested certificate default Duration
	// +optional
	Duration *metav1.Duration `json:"duration,omitempty"`

	// IssuerRef is a reference to the issuer for this CertificateRequest.  If
	// the 'kind' field is not set, or set to 'Issuer', an Issuer resource with
	// the given name in the same namespace as the CertificateRequest will be
	// used.  If the 'kind' field is set to 'ClusterIssuer', a ClusterIssuer with
	// the provided name will be used. The 'name' field in this stanza is
	// required at all times. The group field refers to the API group of the
	// issuer which defaults to 'cert-manager.io' if empty.
	IssuerRef ObjectReference `json:"issuerRef"`

	// Byte slice containing the PEM encoded CertificateSigningRequest
	CSRPEM []byte `json:"csr"`

	// IsCA will mark the resulting certificate as valid for signing. This
	// implies that the 'cert sign' usage is set
	// +optional
	IsCA bool `json:"isCA,omitempty"`

	// Usages is the set of x509 actions that are enabled for a given key.
	// Defaults are ('digital signature', 'key encipherment') if empty
	// +optional
	Usages []KeyUsage `json:"usages,omitempty"`
}

// CertificateStatus defines the observed state of CertificateRequest and
// resulting signed certificate.
type CertificateRequestStatus struct {
	// +optional
	Conditions []CertificateRequestCondition `json:"conditions,omitempty"`

	// Byte slice containing a PEM encoded signed certificate resulting from the
	// given certificate signing request.
	// +optional
	Certificate []byte `json:"certificate,omitempty"`

	// Byte slice containing the PEM encoded certificate authority of the signed
	// certificate.
	// +optional
	CA []byte `json:"ca,omitempty"`

	// FailureTime stores the time that this CertificateRequest failed. This is
	// used to influence garbage collection and back-off.
	// +optional
	FailureTime *metav1.Time `json:"failureTime,omitempty"`
}

// CertificateRequestCondition contains condition information for a CertificateRequest.
type CertificateRequestCondition struct {
	// Type of the condition, currently ('Ready', 'InvalidRequest').
	Type CertificateRequestConditionType `json:"type"`

	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	// +optional
	Message string `json:"message,omitempty"`
}

// CertificateRequestConditionType represents an Certificate condition value.
type CertificateRequestConditionType string

const (
	// CertificateRequestConditionReady indicates that a certificate is ready for use.
	// This is defined as:
	// - The target certificate exists in CertificateRequest.Status
	CertificateRequestConditionReady CertificateRequestConditionType = "Ready"
)
