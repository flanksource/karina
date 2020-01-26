package pgo

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
	"github.com/flanksource/commons/utils"
)

func getGroup(p *platform.Platform) string {
	if strings.HasPrefix(p.PGO.Version, "3.5") {
		return "cr.client-go.k8s.io/v1"
	}
	return "crunchydata.com/v1"
}

func getPgoClient(p *platform.Platform, kind string) (dynamic.NamespaceableResourceInterface, error) {
	client, err := p.GetDynamicClient()
	if err != nil {
		return nil, fmt.Errorf("Error getting dynamic client: %v", err)
	}
	return client.Resource(schema.GroupVersionResource{
		Group:    getGroup(p),
		Version:  "v1",
		Resource: kind,
	}), nil

}

func PSQL(p *platform.Platform, cluster string, sql string) error {
	kubectl := p.GetKubectl()
	return kubectl(" -n pgo exec svc/%s bash -c database -- -c \"psql -c '%s';\"", cluster, sql)
}

func WaitForDB(p *platform.Platform, db string, timeout int) error {
	kubectl := p.GetKubectl()
	start := time.Now()
	for {
		if err := kubectl(" -n pgo exec svc/%s bash -c database -- -c \"psql -c 'SELECT 1';\"", db); err == nil {
			return nil
		}
		if time.Now().Sub(start) > time.Duration(timeout)*time.Second {
			return fmt.Errorf("Timeout waiting for database after %d seconds", time.Now().Sub(start))
		}
		time.Sleep(5 * time.Second)
	}
}

// GetOrCreateDB creates a new postgres cluster and returns access details
func GetOrCreateDB(p *platform.Platform, name string, replicas int, databases ...string) (*types.DB, error) {
	pgo, err := GetPGO(p)
	if err != nil {
		return nil, err
	}

	secret := p.GetSecret(PGO, name+"-postgres-secret")
	var passwd string
	if secret != nil {
		log.Infof("Reusing database %s\n", name)
		passwd = string((*secret)["password"])
	} else {
		log.Infof("Creating new database %s\n", name)
		passwd = utils.RandomString(10)
		if err := pgo("create cluster %s -w %s --replica-count %d --debug", name, passwd, replicas); err != nil {
			return nil, err
		}
	}

	return &types.DB{
		Host:     fmt.Sprintf("%s.%s.svc.cluster.local", name, PGO),
		Username: "postgres",
		Password: passwd,
		Port:     5432,
	}, nil
}
func CreateDatabase(p *platform.Platform, cluster string, databases ...string) error {
	kubectl := p.GetKubectl()
	for _, database := range databases {
		if err := kubectl("-n pgo exec svc/%s bash -c database -- -c \"psql -c 'create database %s'\"", cluster, database); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	}
	return nil
}

func GetPVCName(cluster, db string) string {
	return k8s.GetValidName(fmt.Sprintf("%s-backups-%s", cluster, db))
}

func Backup(p *platform.Platform, cluster, db string) error {
	pvc := GetPVCName(cluster, db)
	if err := p.GetOrCreatePVC("pgo", pvc, "50Gi", "nfs"); err != nil {
		return err
	}
	task := newTask(p, cluster, db, "pgdump", map[string]string{
		"containername": "database",
		"pgdump":        "pgdump",
		"pgdump-all":    "false",
		"pgdump-opts":   "--clean  --if-exists --format d --no-owner --verbose",
		"pvc-name":      pvc,
	})
	return p.Apply("pgo", task)
}

func Restore(p *platform.Platform, cluster, db string) error {
	pvc := GetPVCName(cluster, db)
	task := newTask(p, cluster, db, "pgrestore", map[string]string{
		"pgrestore-from-cluster": cluster,
		"pgrestore-from-pvc":     pvc,
		"pgrestore-opts":         "--clean --exit-on-error --format d --verbose",
		"pvc-name":               pvc,
	})
	return p.Apply("pgo", task)
}

func newTask(p *platform.Platform, cluster, db, task string, params map[string]string) runtime.Object {
	name := fmt.Sprintf("%s-%s-%s", task, cluster, utils.ShortTimestamp())
	if _, ok := params["ccp-image-tag"]; !ok {
		params["ccp-image-tag"] = "centos7-11.4-2.3.3"
	}
	params["pg-cluster"] = cluster
	params[task+"-db"] = db
	params[task] = task
	params[task+"-host"] = cluster
	params[task+"-port"] = "5432"
	params[task+"-user"] = cluster + "-postgres-secret"
	obj := PgoTask{
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: "pgo",
		},
		Spec: map[string]interface{}{
			"namespace":  "pgo",
			"name":       name,
			"parameters": params,
			"tasktype":   task,
		},
	}

	obj.Kind = "Pgtask"
	obj.APIVersion = getGroup(p)
	return obj
}

type PgoTask struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        metav1.ObjectMeta `json:"metadata"`
	Spec            map[string]interface{}
}

func (in PgoTask) DeepCopyObject() runtime.Object {
	return in
}

func (in PgoTask) GetObjectKind() schema.ObjectKind {
	return in.GetObjectKind()
}
