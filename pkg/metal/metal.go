package metal

import (
	metalgo "github.com/metal-pod/metal-go"
)

type Metal struct {
	mclient *metalgo.Driver
}

func New(url, apiToken, hmacKey string) (*Metal, error) {

	driver, err := metalgo.NewDriver(url, apiToken, hmacKey)
	if err != nil {
		return nil, err
	}

	return &Metal{
		mclient: driver,
	}, nil
}

func (m *Metal) NetworkAcquire(nar *metalgo.NetworkAcquireRequest) (*metalgo.NetworkDetailResponse, error) {
	return m.mclient.NetworkAcquire(nar)
}
