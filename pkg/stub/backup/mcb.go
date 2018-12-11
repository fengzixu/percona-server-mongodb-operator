package backup

import (
	"fmt"

	"github.com/Percona-Lab/percona-server-mongodb-operator/pkg/apis/psmdb/v1alpha1"

	motPkg "github.com/percona/mongodb-orchestration-tools/pkg"
	yaml "gopkg.in/yaml.v2"
)

const (
	backupImagePrefix       = "perconalab/mongodb_consistent_backup"
	backupImageVersion      = "1.4.1-3.6"
	backupConfigFile        = "config.yaml"
	backupConfigEnvironment = "production"
)

var (
	backupConfigFileMode = int32(0060)
)

type MCBConfigArchive struct {
	Method v1alpha1.BackupArchiveMethod `yaml:"method,omitempty"`
}

// MCBConfig represents the backup section of the config file for mongodb_consistent_backup
// See: https://github.com/Percona-Lab/mongodb_consistent_backup/blob/master/conf/mongodb-consistent-backup.example.conf#L14
type MCBConfigBackup struct {
	Name     string `yaml:"name"`
	Location string `yaml:"location"`
}

type MCBConfigRotate struct {
	MaxBackups int `yaml:"max_backups,omitempty"`
	MaxDays    int `yaml:"max_days,omitempty"`
}

// MCBConfig represents the config file for mongodb_consistent_backup
// See: https://github.com/Percona-Lab/mongodb_consistent_backup/blob/master/conf/mongodb-consistent-backup.example.conf
type MCBConfig struct {
	Host     string            `yaml:"host"`
	Username string            `yaml:"username,omitempty"`
	Password string            `yaml:"password,omitempty"`
	Archive  *MCBConfigArchive `yaml:"archive,omitempty"`
	Backup   *MCBConfigBackup  `yaml:"backup,omitempty"`
	Rotate   *MCBConfigRotate  `yaml:"rotate,omitempty"`
	Verbose  bool              `yaml:"verbose,omitempty"`
}

func (c *Controller) mongoDBURI(replset *v1alpha1.ReplsetSpec) string {
	return fmt.Sprintf("mongodb+srv://%s-%s.%s.svc.cluster.local/admin?ssl=false&replicaSet=%s",
		c.psmdb.Name,
		replset.Name,
		c.psmdb.Namespace,
		replset.Name,
	)
}

func (c *Controller) newMCBConfigYAML(backup *v1alpha1.BackupSpec, replset *v1alpha1.ReplsetSpec) ([]byte, error) {
	config := &MCBConfig{
		Host:     c.mongoDBURI(replset),
		Username: string(c.usersSecret.Data[motPkg.EnvMongoDBBackupUser]),
		Password: string(c.usersSecret.Data[motPkg.EnvMongoDBBackupPassword]),
		Backup: &MCBConfigBackup{
			Name:     backup.Name,
			Location: "/data/" + c.psmdb.Name,
		},
		Verbose: backup.Verbose,
	}
	if backup.ArchiveMethod != v1alpha1.BackupArchiveMethodNone {
		config.Archive = &MCBConfigArchive{
			Method: backup.ArchiveMethod,
		}
	}
	if backup.Rotate != nil {
		config.Rotate = &MCBConfigRotate{
			MaxBackups: backup.Rotate.MaxBackups,
			MaxDays:    backup.Rotate.MaxDays,
		}
	}
	data := map[string]*MCBConfig{
		backupConfigEnvironment: config,
	}
	return yaml.Marshal(data)
}
