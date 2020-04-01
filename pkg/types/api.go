package types

import (
	"time"

	konfigadm "github.com/moshloop/konfigadm/pkg/types"
)

// Machine represents a running instance of a VM
type Machine interface {
	String() string
	WaitForPoweredOff() error
	GetIP(timeout time.Duration) (string, error)
	WaitForIP() (string, error)
	SetAttributes(attributes map[string]string) error
	GetAttributes() (map[string]string, error)
	Shutdown() error
	PowerOff() error
	Terminate() error
	Name() string
	IP() string
}

type Cluster interface {
	Clone(template VM, config *konfigadm.Config) (Machine, error)
	GetMachine(name string) (Machine, error)
	GetMachines() (map[string]Machine, error)
	GetMachinesByPrefix(prefix string) (map[string]Machine, error)
}
