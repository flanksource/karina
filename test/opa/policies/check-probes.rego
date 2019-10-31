package kubernetes.admission  
  
import data.kubernetes.namespaces  

operations =  {"CREATE", "UPDATE"}
kinds = { "Pod", "Deployment", "ReplicationController", "ReplicaSet", "DaemonSet", "StatefulSet"}

has_key(x, k) { _ = x[k] }
get_container(kind) = input.request.object.spec.template.spec.containers[_] { kinds[kind] }
get_container(kind) = input.request.object.spec.containers[_] { kind == "Pod" }

deny[msg] {
    kinds[input.request.kind.kind]  
    operations[input.request.operation]
    container = get_container(input.request.kind.kind)
    not has_key(container,"readinessProbe")
    msg = "readiness probe is not set"
}

deny[msg] {
    kinds[input.request.kind.kind]  
    operations[input.request.operation]
    container = get_container(input.request.kind.kind)
    not has_key(container,"livenessProbe")
    msg = "liveness  probe is not set"
}