package cmd

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"

	"github.com/fi-ts/cloud-go/api/client"
	"github.com/fi-ts/cloud-go/api/client/version"
	"github.com/fi-ts/cloud-go/api/models"
	mockversion "github.com/fi-ts/cloud-go/test/mocks/version"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/v"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_newVersionCmd(t *testing.T) {
	v.BuildDate = "1.1.1970"
	v.GitSHA1 = "abcdef"
	v.Revision = "v0.0.0"
	v.Version = "v0.0.0"

	mockVersionService := new(mockversion.ClientService)
	cloud := client.CloudAPI{
		Version: mockVersionService,
	}

	var out bytes.Buffer

	mockVersionService.On("Info", mock.Anything, mock.Anything).Return(&version.InfoOK{Payload: &models.RestVersion{Name: pointer.Pointer("cloudctl")}}, nil)
	cfg := &config{
		cloud:           &cloud,
		describePrinter: defaultToYAMLPrinter(&out),
	}
	cmd := newVersionCmd(cfg)

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	mockVersionService.AssertExpectations(t)

	expected := fmt.Sprintf(`---
Client: v0.0.0 (abcdef), v0.0.0, 1.1.1970, %s
Server:
  builddate: null
  gitsha1: null
  min_client_version: null
  name: cloudctl
  revision: null
  version: null
`, runtime.Version())
	assert.Equal(t, expected, string(out.String()))
}
