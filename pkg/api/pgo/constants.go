package pgo

const (
	ANNOTATION_PGHA_BOOTSTRAP_REPLICA    = "pgo-pgha-bootstrap-replica"
	ANNOTATION_CLONE_SOURCE_CLUSTER_NAME = "clone-source-cluster-name"
	ANNOTATION_CLONE_TARGET_CLUSTER_NAME = "clone-target-cluster-name"
	LABEL_NAME                           = "name"
	LABEL_SELECTOR                       = "selector"
	LABEL_OPERATOR                       = "postgres-operator"
	LABEL_PG_CLUSTER                     = "pg-cluster"
	LABEL_PG_CLUSTER_IDENTIFIER          = "pg-cluster-id"
	LABEL_PG_DATABASE                    = "pgo-pg-database"

	LABEL_PGBACKUP = "pgbackup"
	LABEL_PGTASK   = "pg-task"

	LABEL_AUTOFAIL = "autofail"
	LABEL_FAILOVER = "failover"

	LABEL_TARGET = "target"
	LABEL_RMDATA = "pgrmdata"

	LABEL_PGPOLICY          = "pgpolicy"
	LABEL_INGEST            = "ingest"
	LABEL_PGREMOVE          = "pgremove"
	LABEL_PVCNAME           = "pvcname"
	LABEL_COLLECT           = "crunchy_collect"
	LABEL_COLLECT_PG_USER   = "ccp_monitoring"
	LABEL_ARCHIVE           = "archive"
	LABEL_ARCHIVE_TIMEOUT   = "archive-timeout"
	LABEL_CUSTOM_CONFIG     = "custom-config"
	LABEL_NODE_LABEL_KEY    = "NodeLabelKey"
	LABEL_NODE_LABEL_VALUE  = "NodeLabelValue"
	LABEL_REPLICA_NAME      = "replica-name"
	LABEL_CCP_IMAGE_TAG_KEY = "ccp-image-tag"
	LABEL_CCP_IMAGE_KEY     = "ccp-image"
	LABEL_SERVICE_TYPE      = "service-type"
	LABEL_POD_ANTI_AFFINITY = "pg-pod-anti-affinity"
	LABEL_SYNC_REPLICATION  = "sync-replication"

	LABEL_REPLICA_COUNT       = "replica-count"
	LABEL_RESOURCES_CONFIG    = "resources-config"
	LABEL_STORAGE_CONFIG      = "storage-config"
	LABEL_NODE_LABEL          = "node-label"
	LABEL_VERSION             = "version"
	LABEL_PGO_VERSION         = "pgo-version"
	LABEL_UPGRADE_DATE        = "operator-upgrade-date"
	LABEL_DELETE_DATA         = "delete-data"
	LABEL_DELETE_DATA_STARTED = "delete-data-started"
	LABEL_DELETE_BACKUPS      = "delete-backups"
	LABEL_IS_REPLICA          = "is-replica"
	LABEL_IS_BACKUP           = "is-backup"

	LABEL_MINOR_UPGRADE       = "minor-upgrade"
	LABEL_UPGRADE_IN_PROGRESS = "upgrade-in-progress"
	LABEL_UPGRADE_COMPLETED   = "upgrade-complete"
	LABEL_UPGRADE_REPLICA     = "upgrade-replicas"
	LABEL_UPGRADE_PRIMARY     = "upgrade-primary"
	LABEL_UPGRADE_BACKREST    = "upgrade-backrest"

	LABEL_BACKREST                      = "pgo-backrest"
	LABEL_BACKREST_JOB                  = "pgo-backrest-job"
	LABEL_BACKREST_RESTORE              = "pgo-backrest-restore"
	LABEL_CONTAINER_NAME                = "containername"
	LABEL_POD_NAME                      = "podname"
	LABEL_BACKREST_REPO_SECRET          = "backrest-repo-config"
	LABEL_BACKREST_COMMAND              = "backrest-command"
	LABEL_BACKREST_RESTORE_FROM_CLUSTER = "backrest-restore-from-cluster"
	LABEL_BACKREST_RESTORE_TO_PVC       = "backrest-restore-to-pvc"
	LABEL_BACKREST_RESTORE_OPTS         = "backrest-restore-opts"
	LABEL_BACKREST_BACKUP_OPTS          = "backrest-backup-opts"
	LABEL_BACKREST_OPTS                 = "backrest-opts"
	LABEL_BACKREST_PITR_TARGET          = "backrest-pitr-target"
	LABEL_BACKREST_STORAGE_TYPE         = "backrest-storage-type"
	LABEL_BADGER                        = "crunchy-pgbadger"
	LABEL_BADGER_CCPIMAGE               = "crunchy-pgbadger"
	LABEL_BACKUP_TYPE_BASEBACKUP        = "pgbasebackup"
	LABEL_BACKUP_TYPE_BACKREST          = "pgbackrest"
	LABEL_BACKUP_TYPE_PGDUMP            = "pgdump"

	LABEL_PGDUMP_COMMAND = "pgdump"
	LABEL_PGDUMP_RESTORE = "pgdump-restore"
	LABEL_PGDUMP_OPTS    = "pgdump-opts"
	LABEL_PGDUMP_HOST    = "pgdump-host"
	LABEL_PGDUMP_DB      = "pgdump-db"
	LABEL_PGDUMP_USER    = "pgdump-user"
	LABEL_PGDUMP_PORT    = "pgdump-port"
	LABEL_PGDUMP_ALL     = "pgdump-all"
	LABEL_PGDUMP_PVC     = "pgdump-pvc"

	LABEL_RESTORE_TYPE_PGRESTORE = "pgrestore"
	LABEL_PGRESTORE_COMMAND      = "pgrestore"
	LABEL_PGRESTORE_HOST         = "pgrestore-host"
	LABEL_PGRESTORE_DB           = "pgrestore-db"
	LABEL_PGRESTORE_USER         = "pgrestore-user"
	LABEL_PGRESTORE_PORT         = "pgrestore-port"
	LABEL_PGRESTORE_FROM_CLUSTER = "pgrestore-from-cluster"
	LABEL_PGRESTORE_FROM_PVC     = "pgrestore-from-pvc"
	LABEL_PGRESTORE_OPTS         = "pgrestore-opts"
	LABEL_PGRESTORE_PITR_TARGET  = "pgrestore-pitr-target"

	LABEL_PGBASEBACKUP_RESTORE              = "pgo-pgbasebackup-restore"
	LABEL_PGBASEBACKUP_RESTORE_FROM_CLUSTER = "pgbasebackup-restore-from-cluster"
	LABEL_PGBASEBACKUP_RESTORE_FROM_PVC     = "pgbasebackup-restore-from-pvc"
	LABEL_PGBASEBACKUP_RESTORE_TO_PVC       = "pgbasebackup-restore-to-pvc"
	LABEL_PGBASEBACKUP_RESTORE_BACKUP_PATH  = "pgbasebackup-restore-backup-path"

	LABEL_DATA_ROOT   = "data-root"
	LABEL_PVC_NAME    = "pvc-name"
	LABEL_VOLUME_NAME = "volume-name"

	LABEL_SESSION_ID = "sessionid"
	LABEL_USERNAME   = "username"
	LABEL_ROLENAME   = "rolename"
	LABEL_PASSWORD   = "password"

	LABEL_PGBOUNCER                  = "crunchy-pgbouncer"
	LABEL_PGBOUNCER_SECRET           = "pgbouncer-secret"
	LABEL_PGBOUNCER_TASK_ADD         = "pgbouncer-add"
	LABEL_PGBOUNCER_TASK_DELETE      = "pgbouncer-delete"
	LABEL_PGBOUNCER_TASK_CLUSTER     = "pgbouncer-cluster"
	LABEL_PGBOUNCER_TASK_RECONFIGURE = "pgbouncer-reconfigure"
	LABEL_PGBOUNCER_USER             = "pgbouncer-user"
	LABEL_PGBOUNCER_PASS             = "pgbouncer-password"

	LABEL_PGO_LOAD = "pgo-load"

	LABEL_JOB_NAME             = "job-name"
	LABEL_PGBACKREST_STANZA    = "pgbackrest-stanza"
	LABEL_PGBACKREST_DB_PATH   = "pgbackrest-db-path"
	LABEL_PGBACKREST_REPO_PATH = "pgbackrest-repo-path"
	LABEL_PGBACKREST_REPO_HOST = "pgbackrest-repo-host"

	LABEL_PGO_BACKREST_REPO = "pgo-backrest-repo"

	LABEL_PGO_BENCHMARK = "pgo-benchmark"

	// a general label for grouping all the tasks...helps with cleanups
	LABEL_PGO_CLONE = "pgo-clone"

	// the individualized step labels
	LABEL_PGO_CLONE_STEP_1 = "pgo-clone-step-1"
	LABEL_PGO_CLONE_STEP_2 = "pgo-clone-step-2"
	LABEL_PGO_CLONE_STEP_3 = "pgo-clone-step-3"

	LABEL_DEPLOYMENT_NAME = "deployment-name"
	LABEL_SERVICE_NAME    = "service-name"
	LABEL_CURRENT_PRIMARY = "current-primary"

	LABEL_CLAIM_NAME = "claimName"

	LABEL_PGO_PGOUSER = "pgo-pgouser"
	LABEL_PGO_PGOROLE = "pgo-pgorole"
	LABEL_PGOUSER     = "pgouser"
	LABEL_WORKFLOW_ID = "workflowid" // NOTE: this now matches crv1.PgtaskWorkflowID

	LABEL_TRUE  = "true"
	LABEL_FALSE = "false"

	LABEL_NAMESPACE             = "namespace"
	LABEL_PGO_INSTALLATION_NAME = "pgo-installation-name"
	LABEL_VENDOR                = "vendor"
	LABEL_CRUNCHY               = "crunchydata"
	LABEL_PGO_CREATED_BY        = "pgo-created-by"
	LABEL_PGO_UPDATED_BY        = "pgo-updated-by"

	LABEL_PGO_DEFAULT_SC   = "pgo-default-sc"
	LABEL_FAILOVER_STARTED = "failover-started"

	GLOBAL_CUSTOM_CONFIGMAP = "pgo-custom-pg-config"

	LABEL_PGHA_SCOPE             = "crunchy-pgha-scope"
	LABEL_PGHA_DEFAULT_CONFIGMAP = "pgha-default-config"
	LABEL_PGHA_BACKUP_TYPE       = "pgha-backup-type"
	LABEL_PGHA_ROLE              = "role"

	PgclusterResourcePlural = "pgclusters"

	DEFAULT_AUTOFAIL_SLEEP_SECONDS = "30"
	DEFAULT_SERVICE_TYPE           = "ClusterIP"
	LOAD_BALANCER_SERVICE_TYPE     = "LoadBalancer"
	NODEPORT_SERVICE_TYPE          = "NodePort"
	CONFIG_PATH                    = "pgo.yaml"

	DEFAULT_LOG_STATEMENT              = "none"
	DEFAULT_LOG_MIN_DURATION_STATEMENT = "60000"
	DEFAULT_BACKREST_PORT              = 2022
	DEFAULT_PGBADGER_PORT              = "10000"
	DEFAULT_EXPORTER_PORT              = "9187"
	DEFAULT_POSTGRES_PORT              = "5432"
	DEFAULT_PATRONI_PORT               = "8009"
	DEFAULT_BACKREST_SSH_KEY_BITS      = 2048

	DEFAULT_SSH_CONFIG = `
Host *
	StrictHostKeyChecking no
	IdentityFile /tmp/id_rsa
	Port 2022
	User pgbackrest
	`
	DEFAULT_SSHD_CONFIG = `
Port 2022
HostKey /sshd/ssh_host_rsa_key
SyslogFacility AUTHPRIV
PasswordAuthentication no
ChallengeResponseAuthentication yes
PermitRootLogin no
StrictModes no
PubkeyAuthentication yes
AuthorizedKeysFile	/sshd/authorized_keys
UsePAM yes
X11Forwarding no
UsePrivilegeSeparation no
AcceptEnv LANG LC_CTYPE LC_NUMERIC LC_TIME LC_COLLATE LC_MONETARY LC_MESSAGES
AcceptEnv LC_PAPER LC_NAME LC_ADDRESS LC_TELEPHONE LC_MEASUREMENT
AcceptEnv LC_IDENTIFICATION LC_ALL LANGUAGE
AcceptEnv XMODIFIERS
Subsystem	sftp	/usr/libexec/openssh/sftp-server
`
)
