package types

import (
	"time"

	konfigadm "github.com/flanksource/konfigadm/pkg/types"
)

type TagInterface interface {
	GetTags() map[string]string
}

// Machine represents a running instance of a VM
type Machine interface {
	TagInterface
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
	GetAge() time.Duration
	GetTemplate() string
	IP() string
	SetTags(map[string]string) error
}

type NullMachine struct {
	Hostname string
}

func (n NullMachine) String() string {
	return n.Hostname
}
func (n NullMachine) WaitForPoweredOff() error {
	return nil
}
func (n NullMachine) GetIP(timeout time.Duration) (string, error) {
	return "", nil
}
func (n NullMachine) WaitForIP() (string, error) {
	return "", nil
}
func (n NullMachine) SetAttributes(attributes map[string]string) error {
	return nil
}
func (n NullMachine) GetAttributes() (map[string]string, error) {
	return nil, nil
}
func (n NullMachine) Shutdown() error {
	return nil
}
func (n NullMachine) PowerOff() error {
	return nil
}
func (n NullMachine) Terminate() error {
	return nil
}
func (n NullMachine) Name() string {
	return n.Hostname
}
func (n NullMachine) GetAge() time.Duration {
	return 0
}
func (n NullMachine) GetTemplate() string {
	return ""
}
func (n NullMachine) IP() string {
	return ""
}
func (n NullMachine) GetTags() map[string]string {
	return make(map[string]string)
}

func (n NullMachine) SetTags(tags map[string]string) error {
	return nil
}

type Cluster interface {
	Clone(template VM, config *konfigadm.Config) (Machine, error)
	GetMachine(name string) (Machine, error)
	GetMachines() (map[string]Machine, error)
	GetMachinesFor(vm *VM) (map[string]Machine, error)
}
