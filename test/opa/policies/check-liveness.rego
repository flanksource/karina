package kubernetes.admission

import data.kubernetes.namespaces

probes_operations =  {"CREATE", "UPDATE"}
probes_kinds = {"Deployment", "ReplicationController", "ReplicaSet", "DaemonSet", "StatefulSet"}

has_key(x, k) { _ = x[k] }
get_container(kind) = input.request.object.spec.template.spec.containers[_] { probes_kinds[kind] }
get_container(kind) = input.request.object.spec.containers[_] { kind == "Pod" }

deny[msg1] {
    probes_operations[input.request.operation]
    container = get_container(input.request.kind.kind)
    not has_key(container,"livenessProbe")
    msg1 = "liveness probe is not set"
}
