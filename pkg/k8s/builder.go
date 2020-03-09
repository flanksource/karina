package k8s

import (
	apps "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type Builder struct {
	Objects     []runtime.Object
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
}

func (b *Builder) ObjectMeta(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        name,
		Namespace:   b.Namespace,
		Labels:      b.Labels,
		Annotations: b.Annotations,
	}
}

func (b *Builder) Append(objects ...runtime.Object) *Builder {
	Objects := b.Objects
	Objects = append(Objects, objects...)
	b.Objects = Objects
	return b
}

func (b *Builder) AddLabels(labels map[string]string) *Builder {
	b.Labels = labels
	return b
}

func (b *Builder) AddAnnotations(annotations map[string]string) *Builder {
	b.Annotations = annotations
	return b
}

func (b *Builder) SetNamespace(namespace string) *Builder {
	b.Namespace = namespace
	return b
}

func (b *Builder) ConfigMap(name string, data map[string]string) *Builder {
	b.Append(&v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},

		ObjectMeta: b.ObjectMeta(name),
		Data:       data,
	})
	return b
}

func (b *Builder) Secret(name string, data map[string][]byte) *Builder {
	b.Append(&v1.Secret{
		TypeMeta:   metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: b.ObjectMeta(name),
		Type:       "opaque",
		Data:       data,
	})
	return b
}

func (b *Builder) Deployment(name, image string) *DeploymentBuilder {
	return &DeploymentBuilder{
		Builder:     b,
		Name:        name,
		Image:       image,
		replicas:    1,
		resources:   LowResourceRequirements(),
		labels:      b.Labels,
		annotations: b.Annotations,
	}
}

func Deployment(name, image string) *DeploymentBuilder {
	return &DeploymentBuilder{
		Builder:     &Builder{},
		Name:        name,
		Image:       image,
		replicas:    1,
		labels:      make(map[string]string),
		annotations: make(map[string]string),
		resources:   LowResourceRequirements(),
	}
}

type DeploymentBuilder struct {
	Builder         *Builder
	Name, Image, sa string
	replicas        int32
	args            []string
	resources       v1.ResourceRequirements
	volumeMounts    []v1.VolumeMount
	ports           []v1.ContainerPort
	volumes         []v1.Volume
	labels          map[string]string
	annotations     map[string]string
	env             []v1.EnvVar
	nodeAffinity    *v1.Affinity
	podAffinity     *v1.Affinity
	cmd             []string
}

func (d *DeploymentBuilder) Command(cmd ...string) *DeploymentBuilder {
	d.cmd = append(d.cmd, cmd...)
	return d
}

func (d *DeploymentBuilder) EnvVarFromField(env, field string) *DeploymentBuilder {
	d.env = append(d.env, v1.EnvVar{
		Name: env,
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  field,
			},
		},
	})
	return d
}

func (d *DeploymentBuilder) EnvVarFromSecret(env, secret, key string) *DeploymentBuilder {
	d.env = append(d.env, v1.EnvVar{
		Name: env,
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: secret,
				},
				Key: key,
			},
		},
	})
	return d
}

func (d *DeploymentBuilder) EnvVarFromConfigMap(env, configmap, key string) *DeploymentBuilder {
	d.env = append(d.env, v1.EnvVar{
		Name: env,
		ValueFrom: &v1.EnvVarSource{
			ConfigMapKeyRef: &v1.ConfigMapKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: configmap,
				},
				Key: key,
			},
		},
	})
	return d
}

func (d *DeploymentBuilder) AsCronJob(schedule string) *batchv1beta1.CronJob {
	return &batchv1beta1.CronJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1beta1",
			Kind:       "cronjob",
		},
		ObjectMeta: d.Builder.ObjectMeta(d.Name),
		Spec: batchv1beta1.CronJobSpec{
			Schedule: schedule,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: d.PodTemplate(),
				},
			},
			ConcurrencyPolicy: batchv1beta1.ForbidConcurrent,
		},
	}
}

func (d *DeploymentBuilder) PodSpec() v1.PodSpec {
	return v1.PodSpec{
		ServiceAccountName: d.sa,
		Volumes:            d.volumes,
		Containers: []v1.Container{
			{
				Command:         d.cmd,
				Name:            d.Name,
				Image:           d.Image,
				ImagePullPolicy: v1.PullIfNotPresent,
				Ports:           d.ports,
				VolumeMounts:    d.volumeMounts,
				Env:             d.env,
				Args:            d.args,
				Resources:       d.resources,
			},
		},
	}
}

func (d *DeploymentBuilder) ObjectMeta() metav1.ObjectMeta {
	return d.Builder.ObjectMeta(d.Name)
}

func (d *DeploymentBuilder) PodTemplate() v1.PodTemplateSpec {
	return v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: d.GetLabels(),
		},
		Spec: d.PodSpec(),
	}
}

func (d *DeploymentBuilder) AsOneShotJob() *v1.Pod {
	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: d.ObjectMeta(),
		Spec:       d.PodSpec(),
	}
	pod.Spec.RestartPolicy = "Never"
	return &pod
}

func (d *DeploymentBuilder) EnvVars(env map[string]string) *DeploymentBuilder {
	for k, v := range env {
		d.env = append(d.env, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return d
}

func (d *DeploymentBuilder) NodeAffinity(nodeReadinessLabel map[string]string) *DeploymentBuilder {
	matchExpressions := make([]v1.NodeSelectorRequirement, 0)
	if len(nodeReadinessLabel) == 0 {
		return nil
	}
	for k, v := range nodeReadinessLabel {
		matchExpressions = append(matchExpressions, v1.NodeSelectorRequirement{
			Key:      k,
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{v},
		})
	}

	d.nodeAffinity = &v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{{MatchExpressions: matchExpressions}},
			},
		},
	}
	return d
}

func (d *DeploymentBuilder) PodAffinity(labels map[string]string, topologyKey string) *DeploymentBuilder {
	podAffinity := v1.Affinity{
		PodAntiAffinity: &v1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{{
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				TopologyKey: topologyKey,
			}},
		},
	}

	if d.nodeAffinity != nil && d.nodeAffinity.NodeAffinity != nil {
		podAffinity.NodeAffinity = d.nodeAffinity.NodeAffinity
	}

	d.podAffinity = &podAffinity
	return d
}

func (d *DeploymentBuilder) Labels(labels map[string]string) *DeploymentBuilder {
	for k, v := range labels {
		d.labels[k] = v
	}
	return d
}

func (d *DeploymentBuilder) Annotations(annotations map[string]string) *DeploymentBuilder {
	for k, v := range annotations {
		d.annotations[k] = v
	}
	return d
}

func (d *DeploymentBuilder) Args(args ...string) *DeploymentBuilder {
	d.args = args
	return d
}

func (d *DeploymentBuilder) Replicas(replicas int) *DeploymentBuilder {
	d.replicas = int32(replicas)
	return d
}

func (d *DeploymentBuilder) Resources(resources v1.ResourceRequirements) *DeploymentBuilder {
	d.resources = resources
	return d
}

func (d *DeploymentBuilder) MountSecret(secret, path string, mode int32) *DeploymentBuilder {
	d.volumes = append(d.volumes, v1.Volume{
		Name: secret,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName:  secret,
				DefaultMode: &mode,
			},
		},
	})
	d.volumeMounts = append(d.volumeMounts, v1.VolumeMount{
		Name:      secret,
		MountPath: path,
		ReadOnly:  true,
	})

	return d
}

func (d *DeploymentBuilder) MountConfigMap(cm, path string) *DeploymentBuilder {
	d.volumes = append(d.volumes, v1.Volume{
		Name: cm,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: cm,
				},
			},
		},
	})
	d.volumeMounts = append(d.volumeMounts, v1.VolumeMount{
		Name:      cm,
		MountPath: path,
	})
	return d
}

func (d *DeploymentBuilder) ServiceAccount(name string) *DeploymentBuilder {
	d.sa = name
	return d
}

func (d *DeploymentBuilder) GetLabels() map[string]string {
	if len(d.labels) == 0 {
		return map[string]string{
			"name": d.Name,
		}
	} else {
		return d.labels
	}
}

func (d *DeploymentBuilder) Ports(ports ...int32) *DeploymentBuilder {
	for _, port := range ports {
		d.ports = append(d.ports, v1.ContainerPort{
			ContainerPort: port,
			Protocol:      v1.ProtocolTCP,
		})
	}
	return d
}

func (d *DeploymentBuilder) Expose(ports ...int32) *DeploymentBuilder {
	var servicePorts []v1.ServicePort

	for _, port := range ports {
		d.ports = append(d.ports, v1.ContainerPort{
			ContainerPort: port,
		})
		servicePorts = append(servicePorts, v1.ServicePort{
			Port:       port,
			TargetPort: intstr.FromInt(int(port)),
		})
	}
	d.Builder.Append(&v1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "v1"},
		ObjectMeta: d.Builder.ObjectMeta(d.Name),
		Spec: v1.ServiceSpec{
			Selector: d.GetLabels(),
			Ports:    servicePorts,
		},
	})
	return d
}

func (d *DeploymentBuilder) Build() *Builder {
	d.Builder.Append(&apps.Deployment{
		TypeMeta:   metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
		ObjectMeta: d.Builder.ObjectMeta(d.Name),
		Spec: apps.DeploymentSpec{
			Replicas: &d.replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: d.GetLabels(),
			},
			Template: d.PodTemplate(),
		},
	})
	return d.Builder
}

func (b *Builder) ServiceAccount(name string) *ServiceAccountBuilder {
	b.Append(&v1.ServiceAccount{
		TypeMeta:   metav1.TypeMeta{Kind: "ServiceAccount", APIVersion: "v1"},
		ObjectMeta: b.ObjectMeta(name),
	})
	return &ServiceAccountBuilder{
		Builder: b,
		Name:    name,
	}
}

type ServiceAccountBuilder struct {
	*Builder
	Name string
}

func (s *ServiceAccountBuilder) AddRole(role string) *ServiceAccountBuilder {
	s.Append(&rbac.RoleBinding{
		ObjectMeta: s.ObjectMeta(s.Name + "-" + role),
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		Subjects: []rbac.Subject{
			rbac.Subject{
				Name: s.Name,
				Kind: "ServiceAccount",
			},
		},
		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     role,
		},
	})
	return s
}

func (s *ServiceAccountBuilder) AddClusterRole(role string) *ServiceAccountBuilder {
	s.Append(&rbac.ClusterRoleBinding{
		ObjectMeta: s.ObjectMeta(s.Name + "-" + role),
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		Subjects: []rbac.Subject{
			rbac.Subject{
				Name:      s.Name,
				Kind:      "ServiceAccount",
				Namespace: s.Namespace,
			},
		},
		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     role,
		},
	})
	return s
}
