package types

type Boolean bool

type XEnabled struct {
	// +optional
	Disabled Boolean `yaml:"disabled" json:"disabled"`
}

type XDisabled struct {
	// +optional
	Disabled Boolean `yaml:"disabled" json:"disabled"`
	// +optional
	Version string `yaml:"version" json:"version"`
}

func (d XDisabled) IsDisabled() bool {
	if d.Disabled {
		return true
	}
	return d.Version == ""
}
