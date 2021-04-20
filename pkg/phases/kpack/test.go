package kpack

import (
	"github.com/flanksource/commons/console"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/flanksource/kommons"
	"github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test(p *platform.Platform, test *console.TestResults) {
	if p.Kpack.Disabled || &p.Kpack.ImageVersions == nil {
		return
	}

	client, _ := p.GetClientset()
	kommons.TestNamespace(client, Namespace, test)
	kommons.TestDeploy(client, Namespace, "kpack-controller", test)
	if p.E2E{
		TestKpackE2E(p, test)
	}
}

func TestKpackE2E(p *platform.Platform, test *console.TestResults) {
	testName := "kpack-e2e-test"
	if err := p.Apply(Namespace, getDefaultKpackClusterStore()); err != nil {
		test.Failf(testName, "error creating ClusterStore: %v", err)
	}
	if err := p.Apply(Namespace, getDefaultKpackClusterStack()); err != nil {
		test.Failf(testName, "error creating ClusterStack: %v", err)
	}
	if err := p.Apply(Namespace, getDefaultKpackBuilderConfiguration()); err != nil {
		test.Failf(testName, "error creating Builder: %v", err)
	}
	image := getDefaultKpackImageConfiguration()
	if err := p.Apply(Namespace, image); err != nil {
		test.Failf(testName, "error creating Image: %v", err)
	}
	for i := 1; i < 5; i++ {
		var conditionsLength = len(image.Status.Conditions)-1
		if image.Status.Conditions[conditionsLength].Status == "True"{
			test.Passf(testName, "kpack image has been successfully ")
		}
	}
}

func getDefaultKpackClusterStore() *v1alpha1.ClusterStore {
	return &v1alpha1.ClusterStore{
		TypeMeta: metav1.TypeMeta{
			Kind: v1alpha1.ClusterStoreKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "default-cluster-store",
		},
		Spec: v1alpha1.ClusterStoreSpec{
			Sources: []v1alpha1.StoreImage{
				{
					Image: "gcr.io/paketo-buildpacks/java",
				},
				{
					Image: "gcr.io/paketo-buildpacks/nodejs",
				},
			},
		},
	}
}

func getDefaultKpackClusterStack() *v1alpha1.ClusterStack{
	return &v1alpha1.ClusterStack{
		TypeMeta: metav1.TypeMeta{
			Kind: v1alpha1.ClusterStackKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "default-cluster-stack",
		},
		Spec: v1alpha1.ClusterStackSpec{
			Id: "io.buildpacks.stacks.bionic",
			BuildImage: v1alpha1.ClusterStackSpecImage{
				Image: "paketobuildpacks/build:base-cnb",
			},
			RunImage: v1alpha1.ClusterStackSpecImage{
				Image: "paketobuildpacks/run:base-cnb",
			},
		},
	}
}

func getDefaultKpackBuilderConfiguration() *v1alpha1.Builder {
	return &v1alpha1.Builder{
		TypeMeta: metav1.TypeMeta{
			Kind: v1alpha1.BuilderKind,
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "default-builder",
		},
		Spec: v1alpha1.NamespacedBuilderSpec{
			ServiceAccount: "default",
			BuilderSpec: v1alpha1.BuilderSpec{
				Tag: "test-image",
				Stack: v1.ObjectReference{
					Name: "default-cluster-stack",
					Kind: v1alpha1.ClusterStackKind,
				},
				Store: v1.ObjectReference{
					Name: "default-cluster-store",
					Kind: v1alpha1.ClusterStoreKind,
				},
				Order: v1alpha1.Order{
					v1alpha1.OrderEntry{
						Group: []v1alpha1.BuildpackRef{
							{
								BuildpackInfo: v1alpha1.BuildpackInfo{
									Id: "paketo-buildpacks/java",
								},
							},
							{
								BuildpackInfo: v1alpha1.BuildpackInfo{
									Id: "paketo-buildpacks/nodejs",
								},
							},
						},
					},
				},
			},
		},
	}
}

func getDefaultKpackImageConfiguration() *v1alpha1.Image{
	return &v1alpha1.Image{
		TypeMeta: metav1.TypeMeta{
			Kind: "Image",
			APIVersion: "kpack.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "default-image",
		},
		Spec: v1alpha1.ImageSpec{
			Tag: "test-image",
			ServiceAccount: "default",
			Builder: v1.ObjectReference{
				Name: "default-builder",
				Kind: v1alpha1.BuilderKind,
			},
			Source: v1alpha1.SourceConfig{
				Git: &v1alpha1.Git{
					URL: "https://github.com/spring-projects/spring-petclinic",
					Revision: "82cb521d636b282340378d80a6307a08e3d4a4c4",
				},
			},
		},
	}
}