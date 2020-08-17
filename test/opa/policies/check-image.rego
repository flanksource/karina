package kubernetes.admission

import data.kubernetes.namespaces

operations =  {"CREATE", "UPDATE"}
kinds = { "Pod", "Deployment", "ReplicationController", "ReplicaSet", "DaemonSet", "StatefulSet", "Job", "CronJob"}

get_image(kind) = input.request.object.spec.template.spec.containers[_].image { kinds[kind] }
get_image(kind) = input.request.object.spec.containers[_].image { kind == "Pod" }
get_image(kind) = input.request.object.spec.jobTemplate.spec.template.spec.containers[_].image { kind == "CronJob" }

valid_deployment_registries = {registry |
    whitelist := namespaces[input.request.namespace].metadata.annotations["registry-whitelist"]
    registries = split(whitelist, ",")
    registry = registries[_]
}

reg_matches_any(str, patterns) {
    reg_matches(str, patterns[_])
}

reg_matches(str, pattern) {
    contains(str, pattern)
}

deny[msg1] {
    kinds[input.request.kind.kind]
    operations[input.request.operation]
    registry = get_image(input.request.kind.kind)
    not reg_matches_any(registry,valid_deployment_registries)
    msg1 := sprintf("your image registry is not whitelisted:registry=%q", [registry])
}

deny[msg] {
    kinds[input.request.kind.kind]
    operations[input.request.operation]
    image = get_image(input.request.kind.kind)
    not contains(image, ":")
    msg = sprintf("no tag in image-name %q", [image])
}

deny[msg] {
    kinds[input.request.kind.kind]
    operations[input.request.operation]
    image = get_image(input.request.kind.kind)
    [image_name, image_tag] = split(image, ":")
    image_tag = "latest"
    msg = sprintf("invalid image tag — using default latest tag %q", [image])
}
