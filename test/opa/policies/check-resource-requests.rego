package kubernetes.admission  
  
import data.kubernetes.namespaces  

operations =  {"CREATE", "UPDATE"}
kinds = { "Pod", "Deployment", "ReplicationController", "ReplicaSet", "DaemonSet", "StatefulSet"}

get_container(kind) = input.request.object.spec.template.spec.containers[_] { kinds[kind] }
get_container(kind) = input.request.object.spec.containers[_] { kind == "Pod" }


deny[msg] {
    kinds[input.request.kind.kind]  
    operations[input.request.operation]
    container = get_container(input.request.kind.kind)
    count(container.resources) < 2
    msg = sprintf("resource limits OR requests are not set: resources=%q",[container.resources])
}

deny[msg] {
    kinds[input.request.kind.kind]  
    operations[input.request.operation]
    container = get_container(input.request.kind.kind)
    resource_limits = container.resources.limits
    count(resource_limits) < 2
    msg = sprintf("resource limits cpu OR memory are not set: resource_limits=%q",[resource_limits])
}

deny[msg] {
    input.request.kind.kind = "Deployment"
    input.request.operation = "CREATE"
    container = input.request.object.spec.template.spec.containers[_]
    resource_requests = container.resources.requests
    count(resource_requests) < 2
    msg = sprintf("resource requests cpu OR memory are not set: resource_requests=%q",[resource_requests])
}