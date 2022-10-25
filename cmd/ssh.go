package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"strings"
	"time"

	dockerclient "github.com/docker/docker/client"
	"github.com/fi-ts/cloud-go/api/models"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
	"golang.org/x/term"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

const (
	tailscaleImage          = "tailscale/tailscale:v1.32"
	taiscaleStatusRetries   = 50
	proxyConnectionAttempts = 50
)

func (c *config) firewallSSHViaVPN(firewallID string, privateKey []byte, vpn *models.V1VPN) (err error) {
	fmt.Printf("accessing firewall through vpn ")

	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
	if err != nil {
		return fmt.Errorf("failed to initialize Docker client: %w", err)
	}

	// Deploy tailscaled
	ctx := context.Background()
	if err := pullImageIfNotExists(ctx, cli, tailscaleImage); err != nil {
		return fmt.Errorf("failed to pull tailscale image: %w", err)
	}

	socksProxyPort, err := getFreePort()
	if err != nil {
		return err
	}

	containerConfig := &container.Config{
		Image: tailscaleImage,
		Cmd:   []string{"tailscaled", "--tun=userspace-networking", "--no-logs-no-support", fmt.Sprintf("--socks5-server=:%d", socksProxyPort)},
	}
	hostConfig := &container.HostConfig{
		NetworkMode: container.NetworkMode("host"),
		AutoRemove:  true,
	}
	containerName := "tailscaled-" + firewallID
	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return err
	}

	tailscaledContainerID := resp.ID
	if err = cli.ContainerStart(ctx, tailscaledContainerID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	defer func() {
		if e := cli.ContainerStop(ctx, tailscaledContainerID, nil); e != nil {
			if err != nil {
				e = fmt.Errorf("%s: %w", e, err)
			}
			err = e
		}
	}()

	// Exec tailscale up
	execConfig := types.ExecConfig{
		Cmd: []string{"tailscale", "up", "--auth-key=" + *vpn.AuthKey, "--login-server=" + *vpn.Address},
	}
	execResp, err := cli.ContainerExecCreate(ctx, containerName, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create tailscaled exec: %w", err)
	}
	if err := cli.ContainerExecStart(ctx, execResp.ID, types.ExecStartCheck{}); err != nil {
		return fmt.Errorf("failed to start tailscaled exec: %w", err)
	}

	// Connect to the firewall via SSH
	firewallVPNAddr, err := c.getFirewallVPNAddr(ctx, cli, containerName, firewallID)
	if err != nil {
		return fmt.Errorf("failed to get Firewall VPN address: %w", err)
	}
	ip, err := netip.ParseAddr(firewallVPNAddr)
	if err != nil {
		return fmt.Errorf("unable to parse firewall vpn address %w", err)
	}
	addr := ip.String()
	if ip.Is6() {
		addr = fmt.Sprintf("[%s]", ip)
	}

	err = sshClientOverSOCKS5("metal", addr, privateKey, 22, fmt.Sprintf(":%d", socksProxyPort))
	if err != nil {
		return fmt.Errorf("machine console error:%w", err)
	}

	return nil
}

// TailscaleStatus and TailscalePeerStatus structs are used to parse VPN IP for the machine
type TailscaleStatus struct {
	Peer map[string]*TailscalePeerStatus
}

type TailscalePeerStatus struct {
	HostName     string
	TailscaleIPs []string
}

func pullImageIfNotExists(ctx context.Context, cli *dockerclient.Client, tag string) error {
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	for _, i := range images {
		for _, t := range i.RepoTags {
			if t == tag {
				return nil
			}
		}
	}

	reader, err := cli.ImagePull(ctx, tag, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	if _, err := io.Copy(os.Stdout, reader); err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	return nil
}

func (c *config) getFirewallVPNAddr(ctx context.Context, cli *dockerclient.Client, containerName, fwName string) (addr string, err error) {
	// Wait until Peers info is filled
	for i := 0; i < taiscaleStatusRetries; i++ {
		execConfig := types.ExecConfig{
			Cmd:          []string{"tailscale", "status", "--json"},
			AttachStdout: true,
		}
		execResp, err := cli.ContainerExecCreate(ctx, containerName, execConfig)
		if err != nil {
			return "", fmt.Errorf("failed to create tailscale status exec: %w", err)
		}
		resp, err := cli.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
		if err != nil {
			return "", fmt.Errorf("failed to attach to tailscale status exec: %w", err)
		}

		var data string
		s := bufio.NewScanner(resp.Reader)
		for s.Scan() {
			data += s.Text()
		}

		// Skipping noise at the beginning
		var i int
		for _, c := range data {
			if c == '{' {
				break
			}
			i++
		}
		ts := &TailscaleStatus{}
		if err := json.Unmarshal([]byte(data[i:]), ts); err != nil {
			continue
		}

		if ts.Peer != nil {
			for _, p := range ts.Peer {
				if strings.HasPrefix(p.HostName, fwName) {
					return p.TailscaleIPs[0], nil
				}
			}
		}
	}

	return "", fmt.Errorf("failed to find IP for specified firewall")
}

func sshClientOverSOCKS5(user, host string, privateKey []byte, port int, proxyAddr string) error {
	sshConfig, err := getSSHConfig(user, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create SSH config: %w", err)
	}
	sshServerAddress := fmt.Sprintf("%s:%d", host, port)
	client, err := getProxiedSSHClient(sshServerAddress, proxyAddr, sshConfig)
	if err != nil {
		return err
	}

	return createSSHSession(client, nil)
}

func sshClient(user, host string, privateKey []byte, port int, env *env) error {
	fmt.Printf("ssh to %s@%s:%d\n", user, host, port)
	sshConfig, err := getSSHConfig(user, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create SSH config: %w", err)
	}
	sshServerAddress := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", sshServerAddress, sshConfig)
	if err != nil {
		return err
	}

	return createSSHSession(client, env)
}

func getProxiedSSHClient(sshServerAddress, proxyAddr string, sshConfig *ssh.ClientConfig) (*ssh.Client, error) {
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to create a proxy dialer: %w", err)
	}

	var conn net.Conn
	for i := 0; i < proxyConnectionAttempts; i++ {
		fmt.Printf(".")
		conn, err = dialer.Dial("tcp", sshServerAddress)
		if err == nil {
			fmt.Printf("\n")
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy at address %s: %w", proxyAddr, err)
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, sshServerAddress, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh connection: %w", err)
	}

	return ssh.NewClient(c, chans, reqs), nil
}

type env struct {
	key   string
	value string
}

func createSSHSession(client *ssh.Client, env *env) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	if env != nil {
		err = session.Setenv(env.key, env.value)
		if err != nil {
			return err
		}
	}
	// Set IO
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin
	// Set up terminal modes
	// https://net-ssh.github.io/net-ssh/classes/Net/SSH/Connection/Term.html
	// https://www.ietf.org/rfc/rfc4254.txt
	// https://godoc.org/golang.org/x/crypto/ssh
	// THIS IS THE TITLE
	// https://pythonhosted.org/ANSIColors-balises/ANSIColors.html
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,      // enable echoing
		ssh.TTY_OP_ISPEED: 115200, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 115200, // output speed = 14.4kbaud
	}

	fileDescriptor := int(os.Stdin.Fd())

	if term.IsTerminal(fileDescriptor) {
		originalState, err := term.MakeRaw(fileDescriptor)
		if err != nil {
			return err
		}
		defer func() {
			err = term.Restore(fileDescriptor, originalState)
			if err != nil {
				fmt.Printf("error restoring ssh terminal:%v\n", err)
			}
		}()

		termWidth, termHeight, err := term.GetSize(fileDescriptor)
		if err != nil {
			return err
		}

		err = session.RequestPty("xterm-256color", termHeight, termWidth, modes)
		if err != nil {
			return err
		}
	}

	err = session.Shell()
	if err != nil {
		return err
	}

	// You should now be connected via SSH with a fully-interactive terminal
	// This call blocks until the user exits the session (e.g. via CTRL + D)
	return session.Wait()
}

func getSSHConfig(user string, privateKey []byte) (*ssh.ClientConfig, error) {
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		//nolint:gosec
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}, nil
}
func getFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}
