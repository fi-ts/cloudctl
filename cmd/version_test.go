package cmd

import (
	"io"
	"os"
	"testing"

	"github.com/fi-ts/cloud-go/api/client"
	"github.com/fi-ts/cloud-go/api/client/version"
	"github.com/fi-ts/cloud-go/api/models"
	mockversion "github.com/fi-ts/cloud-go/test/mocks/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/pointer"
)

func Test_newVersionCmd(t *testing.T) {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = rescueStdout
	}()

	mockVersionService := new(mockversion.ClientService)
	cloud := client.CloudAPI{
		Version: mockVersionService,
	}

	mockVersionService.On("Info", mock.Anything, mock.Anything).Return(&version.InfoOK{Payload: &models.RestVersion{Name: pointer.StringPtr("cloudctl")}}, nil)
	cfg := &config{
		cloud: &cloud,
	}
	cmd := newVersionCmd(cfg)
	//cmd.SetOut(w)
	//cmd.SetArgs([]string{"--o", "yaml"})
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	mockVersionService.AssertExpectations(t)
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	expected := `client: version not set, please build your app with appropriate ldflags, see https://github.com/metal-stack/v for reference, go1.17
server:
    builddate: null
    gitsha1: null
    name: cloudctl
    revision: null
    version: null
`
	assert.Equal(t, expected, string(out))
}
