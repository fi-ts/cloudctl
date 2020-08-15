package cloud

import (
	"fmt"
	"net/url"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/metal-stack/security"

	"github.com/fi-ts/cloud-go/api/client"
	"github.com/fi-ts/cloud-go/api/client/accounting"
	"github.com/fi-ts/cloud-go/api/client/cluster"
	"github.com/fi-ts/cloud-go/api/client/ip"
	"github.com/fi-ts/cloud-go/api/client/project"
	"github.com/fi-ts/cloud-go/api/client/s3"
	"github.com/fi-ts/cloud-go/api/client/tenant"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// Cloud provides cloud functions
type Cloud struct {
	Cluster     *cluster.Client
	Project     *project.Client
	Tenant      *tenant.Client
	IP          *ip.Client
	Accounting  *accounting.Client
	S3          *s3.Client
	Auth        runtime.ClientAuthInfoWriter
	ConsoleHost string
}

// NewCloud create a new Cloud-Client.
// If specified, hmac takes precedence over apiToken.
func NewCloud(apiurl, apiToken string, hmac string) (*Cloud, error) {

	parsedurl, err := url.Parse(apiurl)
	if err != nil {
		return nil, err
	}
	if parsedurl.Host == "" {
		return nil, fmt.Errorf("invalid url:%s, must be in the form scheme://host[:port]/basepath", apiurl)
	}

	auther := runtime.ClientAuthInfoWriterFunc(func(rq runtime.ClientRequest, rg strfmt.Registry) error {
		if hmac != "" {
			auth := security.NewHMACAuth("Metal-View-All", []byte(hmac))
			auth.AddAuthToClientRequest(rq, time.Now())
			return nil
		}
		if apiToken != "" {
			security.AddUserTokenToClientRequest(rq, apiToken)
			return nil
		}

		return nil
	})

	transport := httptransport.New(parsedurl.Host, parsedurl.Path, []string{parsedurl.Scheme})
	transport.DefaultAuthentication = auther

	cloud := client.New(transport, strfmt.Default)

	c := &Cloud{
		Auth:        auther,
		Cluster:     cloud.Cluster,
		Project:     cloud.Project,
		Tenant:      cloud.Tenant,
		IP:          cloud.IP,
		Accounting:  cloud.Accounting,
		S3:          cloud.S3,
		ConsoleHost: parsedurl.Host,
	}
	return c, nil
}
