package cloud

import (
	"fmt"
	"net/url"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/metal-stack/security"

	"git.f-i-ts.de/cloud-native/cloudctl/api/client"
	"git.f-i-ts.de/cloud-native/cloudctl/api/client/accounting"
	"git.f-i-ts.de/cloud-native/cloudctl/api/client/cluster"
	"git.f-i-ts.de/cloud-native/cloudctl/api/client/ip"
	"git.f-i-ts.de/cloud-native/cloudctl/api/client/project"
	"git.f-i-ts.de/cloud-native/cloudctl/api/client/s3"
	"git.f-i-ts.de/cloud-native/cloudctl/api/client/tenant"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// Cloud provides cloud functions
type Cloud struct {
	Cluster    *cluster.Client
	Project    *project.Client
	Tenant     *tenant.Client
	IP         *ip.Client
	Accounting *accounting.Client
	S3         *s3.Client
	Auth       runtime.ClientAuthInfoWriter
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
		Auth:       auther,
		Cluster:    cloud.Cluster,
		Project:    cloud.Project,
		Tenant:     cloud.Tenant,
		IP:         cloud.IP,
		Accounting: cloud.Accounting,
		S3:         cloud.S3,
	}
	return c, nil
}
