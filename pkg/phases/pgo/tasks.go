package pgo

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/commons/ssh"
	"github.com/flanksource/commons/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	pgoapi "github.com/moshloop/platform-cli/pkg/api/pgo"
	"github.com/moshloop/platform-cli/pkg/k8s"
	"github.com/moshloop/platform-cli/pkg/platform"
	"github.com/moshloop/platform-cli/pkg/types"
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
	secret := p.GetSecret(PGO, name+"-postgres-secret")
	var passwd string
	if secret != nil {
		log.Infof("Reusing database %s\n", name)
		passwd = string((*secret)["password"])
	} else {
		log.Infof("Creating new database %s\n", name)
		passwd = utils.RandomString(10)
		if _, err := CreateCluster(p, name, passwd, replicas); err != nil {
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
	if err := p.GetOrCreatePVC("pgo", pvc, "50Gi", p.PGO.BackupStorage); err != nil {
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
	if p.PGO.DBVersion == "" {
		p.PGO.DBVersion = "12.1"
	}
	name := fmt.Sprintf("%s-%s-%s", task, cluster, utils.ShortTimestamp())
	if _, ok := params["ccp-image-tag"]; !ok {
		params["ccp-image-tag"] = fmt.Sprintf("centos7-%s-%s", p.PGO.DBVersion, p.PGO.Version)
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
	return &obj
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
	return k8s.DynamicKind{
		APIVersion: "crunchydata.com/v1",
		Kind:       "Pgtask",
	}
}

// CreateCluster ...
// pgo create cluster mycluster
func CreateCluster(p *platform.Platform, name string, pass string, replicas int) (*pgoapi.Pgcluster, error) {
	userName := "pgouser"
	spec := pgoapi.PgclusterSpec{}
	spec.PodAntiAffinity = "preferred"
	spec.CCPImage = "crunchy-postgres-ha"
	spec.CCPImageTag = fmt.Sprintf("centos7-12.1-%s", p.PGO.Version)
	spec.Namespace = name
	spec.Name = name
	spec.ClusterName = name
	spec.Port = pgoapi.DEFAULT_POSTGRES_PORT
	spec.PGBadgerPort = pgoapi.DEFAULT_PGBADGER_PORT
	spec.ExporterPort = pgoapi.DEFAULT_EXPORTER_PORT
	spec.SecretFrom = ""
	spec.PrimaryHost = name
	spec.Port = "5432"
	spec.User = ""
	spec.Database = "userdb"
	spec.Replicas = strconv.Itoa(replicas)
	spec.PrimaryHost = name
	spec.PrimarySecretName = fmt.Sprintf("%s%s", name, pgoapi.PrimarySecretSuffix)
	spec.RootSecretName = fmt.Sprintf("%s%s", name, pgoapi.RootSecretSuffix)
	spec.Strategy = "1"
	spec.BackrestS3Endpoint = p.S3.Endpoint
	spec.BackrestS3Bucket = p.PGO.BackupBucket
	spec.BackrestS3Region = p.S3.Region
	spec.BackrestStorage = pgoapi.PgStorageSpec{
		Size:         "50G",
		StorageClass: p.PGO.BackrestStorage,
		StorageType:  "dynamic",
		AccessMode:   "ReadWriteOnce",
	}
	spec.ReplicaStorage = pgoapi.PgStorageSpec{
		Size:         "50G",
		StorageClass: p.PGO.PrimaryStorage,
		StorageType:  "dynamic",
		AccessMode:   "ReadWriteOnce",
	}
	spec.PrimaryStorage = pgoapi.PgStorageSpec{
		Size:         "50G",
		StorageClass: p.PGO.ReplicaStorage,
		StorageType:  "dynamic",
		AccessMode:   "ReadWriteOnce",
		Name:         name,
	}
	spec.User = userName
	spec.UserSecretName = fmt.Sprintf("%s-%s%s", name, userName, pgoapi.UserSecretSuffix)
	spec.UserLabels = map[string]string{
		pgoapi.LABEL_PGO_VERSION:     p.PGO.Version,
		pgoapi.LABEL_ARCHIVE_TIMEOUT: "60",
		pgoapi.LABEL_COLLECT:         "false",
	}
	spec.TablespaceMounts = make(map[string]pgoapi.PgStorageSpec)
	labels := make(map[string]string)

	labels["name"] = name
	labels["deployment-name"] = name
	labels[pgoapi.LABEL_AUTOFAIL] = "true"
	labels[pgoapi.LABEL_BACKREST] = "true"
	labels[pgoapi.LABEL_COLLECT] = "false"
	labels[pgoapi.LABEL_SERVICE_TYPE] = "ClusterIP"
	labels[pgoapi.LABEL_PG_CLUSTER] = name
	labels[pgoapi.LABEL_PGO_VERSION] = p.PGO.Version
	labels[pgoapi.LABEL_ARCHIVE_TIMEOUT] = "60"
	labels[pgoapi.LABEL_BACKREST_STORAGE_TYPE] = p.PGO.BackrestStorage
	labels[pgoapi.LABEL_POD_ANTI_AFFINITY] = "preferred"

	if err := createBackrestRepoSecrets(p, name); err != nil {
		return nil, err
	}
	if err := createSecrets(p, name, userName, pass); err != nil {
		return nil, err
	}
	if err := createWorkflowTask(p, name, userName); err != nil {
		return nil, err
	}

	if err := p.GetOrCreatePVC(Namespace, fmt.Sprintf("%s-pgbr-repo", name), "50Gi", p.PGO.BackrestStorage); err != nil {
		return nil, err
	}
	if err := p.GetOrCreatePVC(Namespace, fmt.Sprintf("%s", name), "50Gi", p.PGO.PrimaryStorage); err != nil {
		return nil, err
	}
	newInstance := &pgoapi.Pgcluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "crunchydata.com/v1",
			Kind:       "Pgcluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: spec,
		Status: pgoapi.PgclusterStatus{
			State:   pgoapi.PgclusterStateCreated,
			Message: "Created, not processed yet",
		},
	}
	if err := p.Apply(Namespace, newInstance); err != nil {
		return nil, err
	}

	return newInstance, nil
}

func createSecrets(p *platform.Platform, clusterName, user string, pass string) error {

	RootSecretName := clusterName + pgoapi.RootSecretSuffix
	PrimarySecretName := clusterName + pgoapi.PrimarySecretSuffix
	UserSecretName := clusterName + "-" + user + pgoapi.UserSecretSuffix

	if err := p.GetOrCreateSecret(RootSecretName, Namespace, map[string][]byte{
		"username": []byte("postgres"),
		"password": []byte(pass),
	}); err != nil {
		return err
	}

	if err := p.GetOrCreateSecret(PrimarySecretName, Namespace, map[string][]byte{
		"username": []byte("primaryuser"),
		"password": []byte(pass),
	}); err != nil {
		return err
	}

	if err := p.GetOrCreateSecret(UserSecretName, Namespace, map[string][]byte{
		"username": []byte(user),
		"password": []byte(pass),
	}); err != nil {
		return err
	}

	return nil
}

func createBackrestRepoSecrets(p *platform.Platform, clusterName string) error {

	keys, err := ssh.NewPrivatePublicKeyPair(pgoapi.DEFAULT_BACKREST_SSH_KEY_BITS)
	if err != nil {
		return err
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", clusterName, pgoapi.LABEL_BACKREST_REPO_SECRET),
			Labels: map[string]string{
				pgoapi.LABEL_VENDOR:            pgoapi.LABEL_CRUNCHY,
				pgoapi.LABEL_PG_CLUSTER:        clusterName,
				pgoapi.LABEL_PGO_BACKREST_REPO: "true",
			},
		},
		Data: map[string][]byte{
			"authorized_keys":         keys.Public,
			"id_rsa":                  keys.Private,
			"ssh_host_rsa_key":        keys.Private,
			"aws-s3-ca.crt":           []byte{},
			"aws-s3-credentials.yaml": []byte(fmt.Sprintf("aws-s3-key: %s\n aws-s3-key-secret: %s\n", p.S3.AccessKey, p.S3.SecretKey)),
			"config":                  []byte(pgoapi.DEFAULT_SSH_CONFIG),
			"sshd_config":             []byte(pgoapi.DEFAULT_SSHD_CONFIG),
		},
	}
	return p.Apply(Namespace, secret)
}

func createWorkflowTask(p *platform.Platform, clusterName, pgouser string) error {
	id, _ := uuid.NewUUID()
	if err := p.Apply(Namespace, &pgoapi.Pgtask{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pgtask",
			APIVersion: "crunchydata.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: Namespace,
			Labels: map[string]string{
				pgoapi.LABEL_PGOUSER:    pgouser,
				pgoapi.LABEL_PG_CLUSTER: clusterName,
				pgoapi.PgtaskWorkflowID: id.String(),
			},
		},
		Spec: pgoapi.PgtaskSpec{
			Namespace: Namespace,
			Name:      clusterName + "-" + pgoapi.PgtaskWorkflowCreateClusterType,
			TaskType:  pgoapi.PgtaskWorkflow,
			Parameters: map[string]string{
				pgoapi.PgtaskWorkflowSubmittedStatus: time.Now().Format(time.RFC3339),
				pgoapi.LABEL_PG_CLUSTER:              clusterName,
				pgoapi.PgtaskWorkflowID:              id.String(),
			},
		},
	}); err != nil {
		return err
	}
	id, _ = uuid.NewUUID()
	if err := p.Apply(Namespace, &pgoapi.Pgtask{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pgtask",
			APIVersion: "crunchydata.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName + "-" + pgoapi.PgtaskBackrestStanzaCreate,
			Namespace: Namespace,
			Labels: map[string]string{
				pgoapi.LABEL_PGOUSER:    pgouser,
				pgoapi.LABEL_PG_CLUSTER: clusterName,
				pgoapi.PgtaskWorkflowID: id.String(),
			},
		},
		Spec: pgoapi.PgtaskSpec{
			Namespace: Namespace,
			Name:      clusterName + "-" + pgoapi.PgtaskBackrestStanzaCreate,
			TaskType:  pgoapi.PgtaskWorkflow,
			Parameters: map[string]string{
				pgoapi.PgtaskWorkflowSubmittedStatus: time.Now().Format(time.RFC3339),
				pgoapi.LABEL_PG_CLUSTER:              clusterName,
				pgoapi.PgtaskWorkflowID:              id.String(),
			},
		},
	}); err != nil {
		return err
	}
	return nil

}
