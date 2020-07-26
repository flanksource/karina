package k8s

import "testing"

type healthFixture struct {
	Name   string
	H1, H2 Health
	Result bool
}

func TestHealthComparisons(t *testing.T) {

	fixtures := []healthFixture{
		{
			Name:   "all healthy",
			H1:     Health{RunningPods: 10},
			H2:     Health{RunningPods: 10},
			Result: false,
		},

		{
			Name:   "less healthy",
			H1:     Health{RunningPods: 8},
			H2:     Health{RunningPods: 10},
			Result: true,
		},
		{
			Name:   "more healthy",
			H1:     Health{RunningPods: 12},
			H2:     Health{RunningPods: 10},
			Result: false,
		},
		{
			Name:   "some unhealthy pods",
			H1:     Health{RunningPods: 10, CrashLoopBackOff: 2},
			H2:     Health{RunningPods: 10},
			Result: true,
		},
		{
			Name:   "more unhealthy pods",
			H1:     Health{RunningPods: 10, CrashLoopBackOff: 4},
			H2:     Health{RunningPods: 10, CrashLoopBackOff: 2},
			Result: true,
		},
		{
			Name:   "same unhealthy pods",
			H1:     Health{RunningPods: 10, CrashLoopBackOff: 2},
			H2:     Health{RunningPods: 10, CrashLoopBackOff: 2},
			Result: false,
		},
	}

	for _, fixture := range fixtures {
		_fixture := fixture
		t.Run(fixture.Name, func(t *testing.T) {
			if _fixture.H1.IsDegradedComparedTo(_fixture.H2, 1) != _fixture.Result {
				t.Fail()
			}
		})
	}
}
