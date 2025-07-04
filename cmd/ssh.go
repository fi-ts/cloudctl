package cmd

import (
	"context"
	"fmt"

	"github.com/fi-ts/cloud-go/api/models"
	metalssh "github.com/metal-stack/metal-lib/pkg/ssh"
	metalvpn "github.com/metal-stack/metal-lib/pkg/vpn"
)

func (c *config) firewallSSHViaVPN(firewallID string, privateKey []byte, vpn *models.V1VPN) (err error) {
	fmt.Printf("accessing firewall through vpn ")
	ctx := context.Background()
	v, err := metalvpn.Connect(ctx, firewallID, *vpn.Address, *vpn.AuthKey)
	if err != nil {
		return err
	}
	defer func() {
		_ = v.Close()
	}()

	opts := []metalssh.ConnectOpt{metalssh.ConnectOptOutputPrivateKey(privateKey)}

	s, err := metalssh.NewClientWithConnection("metal", v.TargetIP, v.Conn, opts...)
	if err != nil {
		return err
	}
	return s.Connect(nil)
}

func (c *config) sshClient(user, host string, privateKey []byte, port int, idToken *string) error {
	opts := []metalssh.ConnectOpt{metalssh.ConnectOptOutputPrivateKey(privateKey)}

	s, err := metalssh.NewClient(user, host, port, opts...)
	if err != nil {
		return err
	}
	var env *metalssh.Env
	if idToken != nil {
		env = &metalssh.Env{"LC_METAL_STACK_OIDC_TOKEN": *idToken}
	}
	return s.Connect(env)
}
