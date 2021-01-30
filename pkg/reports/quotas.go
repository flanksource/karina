package reports

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/flanksource/kommons"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ReportOptions struct {
	Path        string
	Annotations []string
	Format      string
}

func GetNamespaces(specs kommons.Specs) map[string]map[string]string {
	namespaces := map[string]map[string]string{}

	for _, ns := range specs.FilterBy("Namespace") {
		namespaces[ns.GetName()] = ns.GetAnnotations()
	}
	return namespaces
}

func Quotas(opts ReportOptions) error {
	specs, err := kommons.Walk(opts.Path)
	if err != nil {
		return err
	}

	// Standard namespaces
	namespaces := GetNamespaces(specs)
	w := tabwriter.NewWriter(os.Stdout, 3, 2, 3, ' ', tabwriter.DiscardEmptyColumns)
	SEP := "\t"
	if opts.Format == "csv" {
		SEP = ";"
	}
	fmt.Fprintf(w, "TYPE%sNAMESPACE%s", SEP, SEP)
	for _, annotation := range opts.Annotations {
		fmt.Fprintf(w, "%s%s", strings.ToUpper(annotation), SEP)
	}
	fmt.Fprintf(w, "LIMIT_GB%s\n", SEP)
	for _, quota := range specs.FilterBy("ResourceQuota") {
		fmt.Fprintf(w, "ResourceQuota%s", SEP)
		limit, found, err := unstructured.NestedString(quota.Object, "spec", "hard", "limits", "memory")
		if err != nil {
			return err
		}

		if !found {
			limit, _, _ = unstructured.NestedString(quota.Object, "spec", "hard", "limits.memory")
		}

		annotations := namespaces[quota.GetNamespace()]
		fmt.Fprintf(w, "%s%s", quota.GetNamespace(), SEP)
		for _, annotation := range opts.Annotations {
			fmt.Fprintf(w, "%s%s", annotations[annotation], SEP)
		}
		fmt.Fprintf(w, "%d%s\n", normalizeGB(limit), SEP)
	}

	// ClusterResourceQuota
	for _, quota := range specs.FilterBy("ClusterResourceQuota") {
		fmt.Fprintf(w, "ClusterResourceQuota%s", SEP)
		limit, found, err := unstructured.NestedString(quota.Object, "spec", "hard", "limits", "memory")
		if err != nil {
			return err
		}

		if !found {
			limit, _, _ = unstructured.NestedString(quota.Object, "spec", "hard", "limits.memory")
		}
		annotations := quota.GetAnnotations()
		fmt.Fprintf(w, "%s%s", quota.GetName(), SEP)
		for _, annotation := range opts.Annotations {
			fmt.Fprintf(w, "%s%s", annotations[annotation], SEP)
		}
		fmt.Fprintf(w, "%d%s\n", normalizeGB(limit), SEP)
	}

	for _, typ := range []string{"PostgresqlDB", "NamespaceRequest"} {
		for _, instance := range specs.FilterBy(typ) {
			memory, _, err := unstructured.NestedString(instance.Object, "spec", "memory")
			if err != nil {
				_memory, found, err := unstructured.NestedInt64(instance.Object, "spec", "memory")
				if found && err == nil {
					memory = fmt.Sprintf("%d", _memory)
				} else {
					return err
				}
			}

			replicas, _, _ := unstructured.NestedInt64(instance.Object, "spec", "replicas")

			if replicas == 0 {
				replicas = 1
			}
			annotations := instance.GetAnnotations()
			fmt.Fprintf(w, "%s%s%s%s", typ, SEP, instance.GetName(), SEP)
			for _, annotation := range opts.Annotations {
				if val, ok := annotations[annotation]; ok {
					fmt.Fprintf(w, "%s%s", val, SEP)
				} else {
					val, _, _ := unstructured.NestedString(instance.Object, "spec", annotation)
					fmt.Fprintf(w, "%s%s", val, SEP)
				}
			}
			fmt.Fprintf(w, "%d%s\n", replicas*normalizeGB(memory), SEP)
		}
	}
	w.Flush()
	return nil
}

// normalizeGB converts size values like 100Gi into number of GB e.g. 100
func normalizeGB(size string) int64 {
	if val, err := strconv.Atoi(size); err == nil {
		return int64(val)
	}
	qty, err := resource.ParseQuantity(size)
	if err != nil {
		return 0
	}
	return qty.Value() / 1024 / 1024 / 1024
}
