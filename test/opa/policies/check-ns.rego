package kubernetes.admission  
  
import data.kubernetes.namespaces
import data.discovery.ca

has_key(obj, k) { _ = obj[k]}
has_item(list, i) {list[_] = i}

deny[msg] {
    input.request.kind.kind = "Namespace"
    input.request.operation = "CREATE"
    metadata = input.request.object.metadata
    not has_key(metadata,"labels")
    msg = "No company lables specified!"
}

deny[msg] {
    input.request.kind.kind = "Namespace"
    input.request.operation = "CREATE"
    incoming_service = input.request.object.metadata.labels.service
    incoming_company = input.request.object.metadata.labels.company
    labels = input.request.object.metadata.labels
    not has_key(ca.companies,incoming_company )
    msg = sprintf("Your company name does not exist company: %q",[incoming_company])
}

deny[msg] {
    input.request.kind.kind = "Namespace"
    input.request.operation = "CREATE"
    incoming_service = input.request.object.metadata.labels.service
    incoming_company = input.request.object.metadata.labels.company
    services = ca.companies[incoming_company]
    labels = input.request.object.metadata.labels
    not has_item(services, incoming_service)
    msg = sprintf("Your company cannot use this service: %q",[incoming_service])
}