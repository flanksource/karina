package operator

import (
	"errors"

	karinav1 "github.com/flanksource/karina/pkg/api/operator/v1"
)

type Operator struct {
}

func New() (*Operator, error) {
	_ = karinav1.KarinaConfig{}
	return &Operator{}, nil
}

func (o *Operator) Run() error {
	return errors.New("Not implemented")
}
