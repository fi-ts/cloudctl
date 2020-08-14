package helper

import (
	"fmt"
	"os"
	"os/user"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/crypto/ssh/terminal"
)

// SSHClient opens a interactive ssh session to the host on port with user, authenticated by the key.
func SSHClient(username, host string, port int, privateKey []byte) error {
	publicKeyAuthMethod, err := publicKey(privateKey)
	if err != nil {
		return err
	}
	user, err := user.Current()
	if err != nil {
		return err
	}
	// from https://skarlso.github.io/2019/02/17/go-ssh-with-host-key-verification/
	// TODO: still complains if no known_hosts entry exists
	// see: https://github.com/golang/crypto/blob/master/ssh/example_test.go
	// and: https://stackoverflow.com/questions/45441735/ssh-handshake-complains-about-missing-host-key
	hostKeyCallback, err := knownhosts.New(user.HomeDir + "/.ssh/known_hosts")
	if err != nil {
		return fmt.Errorf("could not create hostkeycallback function: %v", err)
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			publicKeyAuthMethod,
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         2 * time.Second,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	// TODO for machine access we need agent forwarding
	// agent.RequestAgentForwarding(session)

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

	if terminal.IsTerminal(fileDescriptor) {
		originalState, err := terminal.MakeRaw(fileDescriptor)
		if err != nil {
			return err
		}
		defer terminal.Restore(fileDescriptor, originalState)

		termWidth, termHeight, err := terminal.GetSize(fileDescriptor)
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

func publicKey(privateKey []byte) (ssh.AuthMethod, error) {
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}
