package cmd

import (
	"fmt"
	"io"
	"os"
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
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = rescueStdout
	}()

	v.BuildDate = "1.1.1970"
	v.GitSHA1 = "abcdef"
	v.Revision = "v0.0.0"
	v.Version = "v0.0.0"

	mockVersionService := new(mockversion.ClientService)
	cloud := client.CloudAPI{
		Version: mockVersionService,
	}

	mockVersionService.On("Info", mock.Anything, mock.Anything).Return(&version.InfoOK{Payload: &models.RestVersion{Name: pointer.Pointer("cloudctl")}}, nil)
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

	expected := fmt.Sprintf(`client: v0.0.0 (abcdef), v0.0.0, 1.1.1970, %s
server:
    builddate: null
    gitsha1: null
    minclientversion: null
    name: cloudctl
    revision: null
    version: null
`, runtime.Version())
	assert.Equal(t, expected, string(out))
}
