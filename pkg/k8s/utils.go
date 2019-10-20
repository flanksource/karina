package k8s

import "strings"

func GetValidName(name string) string {
	return strings.ReplaceAll(name, "_", "-")
}
