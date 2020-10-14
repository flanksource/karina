package reports

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/flanksource/karina/pkg/k8s"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ReportOptions struct {
	Path        string
	Annotations []string
	Format      string
}

func GetNamespaces(specs k8s.Specs) map[string]map[string]string {
	namespaces := map[string]map[string]string{}

	for _, ns := range specs.FilterBy("Namespace") {
		namespaces[ns.GetName()] = ns.GetAnnotations()
	}
	return namespaces
}

func Quotas(opts ReportOptions) error {
	specs, err := k8s.Walk(opts.Path)
	if err != nil {
		return err
	}
	namespaces := GetNamespaces(specs)
	w := tabwriter.NewWriter(os.Stdout, 3, 2, 3, ' ', tabwriter.DiscardEmptyColumns)
	SEP := "\t"
	if opts.Format == "csv" {
		SEP = ";"
	}
	fmt.Fprintf(w, "NAMESPACE%s", SEP)
	for _, annotation := range opts.Annotations {
		fmt.Fprintf(w, "%s%s", strings.ToUpper(annotation), SEP)
	}
	fmt.Fprintf(w, "LIMIT_GB%s\n", SEP)
	for _, quota := range specs.FilterBy("ResourceQuota") {
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
	for _, quota := range specs.FilterBy("ClusterResourceQuota") {
		limit, found, err := unstructured.NestedString(quota.Object, "spec", "hard", "limits", "memory")
		if err != nil {
			return err
		}

		if !found {
			limit, _, _ = unstructured.NestedString(quota.Object, "spec", "hard", "limits.memory")
		}
		annotations := quota.GetAnnotations()
		fmt.Fprintf(w, "dynamic%s", SEP)
		for _, annotation := range opts.Annotations {
			fmt.Fprintf(w, "%s%s", annotations[annotation], SEP)
		}
		fmt.Fprintf(w, "%d%s\n", normalizeGB(limit), SEP)
	}
	w.Flush()
	return nil
}

// normalizeGB converts size values like 100Gi into number of GB e.g. 100
func normalizeGB(size string) int64 {
	qty, err := resource.ParseQuantity(size)
	if err != nil {
		return 0
	}
	return qty.Value() / 1024 / 1024 / 1024
}
