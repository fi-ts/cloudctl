package cmd

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/google/uuid"
	"github.com/tailscale/golang-x-crypto/ssh"

	"tailscale.com/tsnet"

	"github.com/fi-ts/cloud-go/api/models"
	"golang.org/x/term"
)

func (c *config) firewallSSHViaVPN(firewallID string, privateKey []byte, vpn *models.V1VPN) (err error) {
	fmt.Printf("accessing firewall through vpn ")
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	randomSuffix, _, _ := strings.Cut(uuid.NewString(), "-")
	hostname = fmt.Sprintf("cloudctl-%s-%s", hostname, randomSuffix)
	tempDir, err := os.MkdirTemp("", hostname)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	s := &tsnet.Server{
		Hostname:   hostname,
		ControlURL: *vpn.Address,
		AuthKey:    *vpn.AuthKey,
		Dir:        tempDir,
	}
	defer s.Close()

	// now disable logging, maybe altogether later
	if os.Getenv("DEBUG") == "" {
		s.Logf = func(format string, args ...any) {}
	}

	start := time.Now()
	lc, err := s.LocalClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	var firewallVPNIP netip.Addr
	err = retry.Do(
		func() error {
			fmt.Printf(".")
			status, err := lc.Status(ctx)
			if err != nil {
				return err
			}
			if status.Self.Online {
				for _, peer := range status.Peer {
					if strings.HasPrefix(peer.HostName, firewallID) {
						firewallVPNIP = peer.TailscaleIPs[0]
						fmt.Printf(" connected to %s (ip %s) took: %s\n", firewallID, firewallVPNIP, time.Since(start))
						return nil
					}
				}
			}
			return fmt.Errorf("did not get online")
		},
		retry.Attempts(50),
	)
	if err != nil {
		return err
	}
	// disable logging after successful connect
	s.Logf = func(format string, args ...any) {}

	conn, err := lc.DialTCP(ctx, firewallVPNIP.String(), 22)
	if err != nil {
		return err
	}

	return sshClientWithConn("metal", hostname, privateKey, conn)
}

// sshClient opens an interactive ssh session to the host on port with user, authenticated by the key.
func sshClientWithConn(user, host string, privateKey []byte, conn net.Conn) error {
	sshConfig, err := getSSHConfig(user, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create SSH config: %w", err)
	}

	sshConn, sshChan, req, err := ssh.NewClientConn(conn, host, sshConfig)
	if err != nil {
		return err
	}
	client := ssh.NewClient(sshConn, sshChan, req)
	if err != nil {
		return err
	}
	defer client.Close()

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
