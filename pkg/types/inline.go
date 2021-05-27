package types

type Boolean bool

type XEnabled struct {
	// +optional
	Disabled bool `yaml:"disabled" json:"disabled"`
}

type XDisabled struct {
	// +optional
	Disabled string `yaml:"disabled" json:"disabled"`
	// +optional
	Version string `yaml:"version" json:"version"`
}

func (d XDisabled) IsDisabled() bool {
	if d.Disabled == "true" {
		return true
	}
	return d.Version == ""
}
