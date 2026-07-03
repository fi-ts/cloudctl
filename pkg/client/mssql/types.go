package mssql

type Database struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Project          string `json:"project"`
	Version          string `json:"version"`
	StorageGB        int    `json:"storage_gb"`
	Host             string `json:"host"`
	Port             int    `json:"port"`
	Status           string `json:"status"`
	CreatedTimestamp string `json:"created_timestamp"`
}

var AllowedVersions = []string{"2019", "2022"}

const (
	StatusPending  = "pending"
	StatusRunning  = "running"
	StatusStopped  = "stopped"
	StatusFailed   = "failed"
	DefaultPort    = 1433
	DefaultStorage = 20
)

func IsAllowedVersion(version string) bool {
	for _, v := range AllowedVersions {
		if v == version {
			return true
		}
	}
	return false
}
