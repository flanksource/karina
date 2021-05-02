package v1

type PodMutaterConfig struct {
	AnnotationsMap         map[string]bool
	Annotations            []string
	RegistryWhitelist      []string
	DefaultRegistryPrefix  string
	DefaultImagePullSecret string
	TolerationsAnnotation  string
}
