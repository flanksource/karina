package reports

import (
	"context"
	"fmt"
	"github.com/flanksource/karina/pkg/platform"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

func Violations(platform *platform.Platform) error {
	k8sclient, err := platform.GetClientset()
	if err != nil {
		return errors.Wrap(err, "Failed to create clientset")
	}

	_, resources, err := k8sclient.ServerGroupsAndResources()
	if err != nil {
		return errors.Wrap(err, "Failed to obtain server resources")
	}

	var constraintTemplates []v1.APIResource
	apiString := "/apis/"
	for _, res := range resources {
		if strings.HasPrefix(res.GroupVersion, "constraints.gatekeeper.sh") {
			constraintTemplates = res.APIResources
			apiString = apiString + res.GroupVersion
			break
		}
	}

	if len(constraintTemplates) == 0 {
		return errors.New("Cluster has no gatekeeper constraints")
	}

	w := tabwriter.NewWriter(os.Stdout, 3, 2, 3, ' ', tabwriter.DiscardEmptyColumns)
	fmt.Fprintf(w, "NAME\tCONSTRAINT\tACTION\tVIOLATIONS\tAUDIT\n")
	for _, ct := range constraintTemplates {
		if ct.Categories != nil {
			constraintsList := &unstructured.UnstructuredList{}
			err := k8sclient.RESTClient().
				Get().
				AbsPath(apiString, "/", ct.Name).
				Timeout(32 * time.Second).
				Do(context.Background()).
				Into(constraintsList)
			if err != nil {
				return errors.Wrapf(err, "Could not retrieve %v objects", ct.Name)
			}
			for _, constraint := range constraintsList.Items {
				name, _, err := unstructured.NestedString(constraint.Object, "metadata", "name")
				if err != nil {
					fmt.Fprintf(w, "\t")
				} else {
					fmt.Fprintf(w, "%s\t", name)
				}
				kind, _, err := unstructured.NestedString(constraint.Object, "kind")
				if err != nil {
					fmt.Fprintf(w, "\t")
				} else {
					fmt.Fprintf(w, "%s\t", kind)
				}
				enforcement, _, err := unstructured.NestedString(constraint.Object, "spec", "enforcementAction")
				if err != nil {
					fmt.Fprintf(w, "\t")
				} else {
					fmt.Fprintf(w, "%s\t", enforcement)
				}
				vcount, _, err := unstructured.NestedInt64(constraint.Object, "status", "totalViolations")
				if err != nil {
					fmt.Fprintf(w, "\t")
				} else {
					fmt.Fprintf(w, "%v\t", vcount)
				}
				atime, _, err := unstructured.NestedString(constraint.Object, "status", "auditTimestamp")
				if err != nil {
					fmt.Fprintf(w, "\t")
				} else {
					fmt.Fprintf(w, "%s\t", atime)
				}
			}
			fmt.Fprint(w, "\n")
		}
	}

	_ = w.Flush()
	return nil
}
