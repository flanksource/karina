package kubernetes.admission  
  
import data.kubernetes.namespaces  

has_key(x, k) { _ = x[k] }

deny[msg] {
    input.request.kind.kind = "Namespace"
    input.request.operation = "CREATE"
    metadata = input.request.object.metadata
    not has_key(metadata,"labels")
    msg = "no lables specified"
}

deny[msg] {
    input.request.kind.kind = "Namespace"
    input.request.operation = "CREATE"
    labels = input.request.object.metadata.labels
    not has_key(labels,"team-name")
    msg = sprintf("please add team-name label. your labels are: %q",[labels])
}
