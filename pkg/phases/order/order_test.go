package order

import (
	"reflect"
	"testing"
)

func TestMergePhases(t *testing.T) {
	var PhasesOne = map[Cmd]Phase{
		ArgoRollouts: makePhase(nil, ArgoRollouts),
		ArgoOperator: makePhase(nil, ArgoOperator),
		Auditbeat:    makePhase(nil, Auditbeat),
		BootstrapCmd: makePhase(nil, BootstrapCmd),
	}

	var PhasesTwo = map[Cmd]Phase{
		Apacheds:     makePhase(nil, Apacheds),
		Antrea:       makePhase(nil, Antrea),
		Base:         makePhase(nil, Base),
		BootstrapCmd: makePhase(nil, BootstrapCmd),
	}

	var PhasesThree = map[Cmd]Phase{
		BootstrapCmd:    makePhase(nil, BootstrapCmd),
		Vsphere:         makePhase(nil, Vsphere),
		CloudController: makePhase(nil, CloudController),
		StubsCmd:        makePhase(nil, StubsCmd),
	}

	var DesiredMap = map[Cmd]Phase{
		Vsphere:         makePhase(nil, Vsphere),
		CloudController: makePhase(nil, CloudController),
		StubsCmd:        makePhase(nil, StubsCmd),
		ArgoRollouts:    makePhase(nil, ArgoRollouts),
		ArgoOperator:    makePhase(nil, ArgoOperator),
		Auditbeat:       makePhase(nil, Auditbeat),
		Apacheds:        makePhase(nil, Apacheds),
		Antrea:          makePhase(nil, Antrea),
		Base:            makePhase(nil, Base),
		BootstrapCmd:    makePhase(nil, BootstrapCmd),
	}

	var AllPhases = []map[Cmd]Phase{PhasesOne, PhasesTwo, PhasesThree}

	type args struct {
		phaseMaps []map[Cmd]Phase
	}
	var tests = []struct {
		name string
		args args
		want map[Cmd]Phase
	}{
		{
			name: "Testing phase merge with multiple phase maps",
			args: args{
				phaseMaps: AllPhases,
			},
			want: DesiredMap,
		},
		{
			name: "Testing phase merge with a single input",
			args: args{
				phaseMaps: []map[Cmd]Phase{PhasesOne},
			},
			want: PhasesOne,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergePhases(tt.args.phaseMaps...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergePhases() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakePhase(t *testing.T) {
	type args struct {
		df   DeployFn
		name Cmd
	}
	tests := []struct {
		name string
		args args
		want Phase
	}{
		{
			name: "Basic",
			args: args{df: nil, name: ArgoRollouts},
			want: Phase{Fn: nil, Name: "argo-rollout"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makePhase(tt.args.df, tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makePhase() = %v, want %v", got, tt.want)
			}
		})
	}
}
