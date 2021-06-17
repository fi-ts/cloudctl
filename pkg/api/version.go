package api

import (
	cloudmodels "github.com/fi-ts/cloud-go/api/models"
)

type Version struct {
	Client string                   `yaml:"client"`
	Server *cloudmodels.RestVersion `yaml:"server,omitempty"`
}
