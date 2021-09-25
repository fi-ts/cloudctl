package cmd

import (
	"bytes"
	"io"
	"testing"

	"github.com/fi-ts/cloud-go/api/client"
	"github.com/fi-ts/cloud-go/api/client/version"
	"github.com/fi-ts/cloud-go/api/models"
	mockversion "github.com/fi-ts/cloud-go/test/mocks/version"
	"github.com/fi-ts/cloudctl/cmd/output"
	"github.com/fi-ts/cloudctl/pkg/api"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v3"
	"k8s.io/utils/pointer"
)

func Test_newVersionCmd(t *testing.T) {
	mockVersionService := new(mockversion.ClientService)
	cloud := client.CloudAPI{
		Version: mockVersionService,
	}

	mockVersionService.On("Info", mock.Anything, mock.Anything).Return(&version.InfoOK{Payload: &models.RestVersion{Name: pointer.StringPtr("cloudctl")}}, nil)
	b := bytes.NewBufferString("")
	printer, err := output.NewPrinter("yaml", "", "", true, b)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config{
		cloud:   &cloud,
		printer: printer,
	}
	cmd := newVersionCmd(cfg)
	// cmd.SetOut(b)
	// cmd.SetArgs([]string{"hi-via-args"}) not needed here
	err = cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	mockVersionService.AssertExpectations(t)

	out, err := io.ReadAll(b)

	if err != nil {
		t.Fatal(err)
	}
	var apiVersion api.Version
	err = yaml.Unmarshal(out, &apiVersion)
	if err != nil {
		t.Fatal(err)
	}

	if apiVersion.Client == "" {
		t.Fatalf("Expected client version to be set")
	}
	if apiVersion.Server == nil {
		t.Fatal("Expected server")
	}
	if *apiVersion.Server.Name != "cloudctl" {
		t.Fatalf("Expected cloudctl, got:%s", *apiVersion.Server.Name)
	}

}
