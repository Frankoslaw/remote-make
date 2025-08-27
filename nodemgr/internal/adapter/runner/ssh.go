package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"nodemgr/internal/core/domain"
	"nodemgr/internal/core/port"
	"nodemgr/internal/core/service"
	"os"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SSHRunner struct {
	client *ssh.Client
}

func NewSSHRunner(node port.NodeClient) (*SSHRunner, error) {
	if !node.HasCap("ssh.enabled") {
		return nil, errors.New("node does not support ssh")
	}

	conn, err := node.Conn()
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: conn.SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(conn.SSHPass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),

		Timeout: 10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", conn.PublicIP, conn.SSHPort)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", addr, err)
	}

	return &SSHRunner{client: client}, nil
}

func (l *SSHRunner) ID() string { return "ssh.enabled" }

func (l *SSHRunner) Attach(ctx context.Context, cmd []string, stdin io.Reader, stdout, stderr io.Writer) (func() error, error) {
	if l.client == nil {
		return nil, errors.New("ssh client is not initialized")
	}

	session, err := l.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	stdinPipe, err := session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("unable to setup stdin for session: %w", err)
	}

	session.Stdout = stdout
	session.Stderr = stderr

	if err := session.Start(strings.Join(cmd, " ")); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	go func() {
		if stdin != nil {
			_, _ = io.Copy(stdinPipe, stdin)
		}
		_ = stdinPipe.Close()
	}()

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():

			_ = session.Close()
		case <-done:

		}
	}()

	waitFunc := func() error {
		defer close(done)
		defer session.Close()
		if err := session.Wait(); err != nil {

			return err
		}
		return nil
	}

	return waitFunc, nil
}

func (l *SSHRunner) Copy(ctx context.Context, src, dst string) error {
	if l.client == nil {
		return errors.New("ssh client is not initialized")
	}

	sftpClient, err := sftp.NewClient(l.client)
	if err != nil {
		return fmt.Errorf("failed to create sftp client: %w", err)
	}
	defer sftpClient.Close()

	if _, err := sftpClient.Stat(src); err == nil {

		srcFile, err := sftpClient.Open(src)
		if err != nil {
			return fmt.Errorf("failed to open remote source file: %w", err)
		}
		defer srcFile.Close()

		localFile, err := os.Create(dst)
		if err != nil {
			return fmt.Errorf("failed to create local destination file: %w", err)
		}
		defer localFile.Close()

		_, err = io.Copy(localFile, srcFile)
		return err
	}

	localFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open local source file: %w", err)
	}
	defer localFile.Close()

	dstFile, err := sftpClient.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create remote destination file: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, localFile)
	return err
}

func (l *SSHRunner) Exec(ctx context.Context, cmd domain.Command) (domain.ExecResult, error) {
	return service.DefaultExec(ctx, l, cmd)
}

var _ port.Runner = (*SSHRunner)(nil)
