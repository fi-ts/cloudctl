package api

import (
	cloudmodels "github.com/fi-ts/cloud-go/api/models"
)

type Version struct {
	Client string                   `json:"client" yaml:"client"`
	Server *cloudmodels.RestVersion `json:"server,omitempty" yaml:"server,omitempty"`
}
