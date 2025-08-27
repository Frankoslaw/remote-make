package main

import (
	"context"
	"log"
	"nodemgr/internal/adapter/provisioner"
	"nodemgr/internal/adapter/runner"
	"nodemgr/internal/core/domain"
	"time"
)

func main() {
	log.Printf("Starting Node Manager...")

	localProvisioner := provisioner.NewLocalProvisioner()
	log.Printf("Using provisioner: %s", localProvisioner.ID())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	localNode, err := localProvisioner.Up(ctx, map[string]string{})
	if err != nil {
		log.Fatalf("failed to provision local node: %v", err)
	}
	conn, err := localNode.Conn()
	if err != nil {
		log.Fatalf("failed to get local node connection info: %v", err)
	}

	localNode.SetCap("ssh.enabled", true)
	conn.PublicIP = "raspberrypi.local"
	conn.SSHUser = "frankoslaw"
	conn.SSHPass = "Lolalola453"
	conn.SSHPort = 22

	sshRunner, err := runner.NewSSHRunner(localNode)
	log.Printf("Using runner: %s", sshRunner.ID())

	if err != nil {
		log.Fatalf("failed to create SSH runner: %v", err)
	}
	localRunner := sshRunner

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := domain.Command{
		Args:  []string{"uname -a"},
		Stdin: []byte(""),
	}
	res, err := localRunner.Exec(ctx, cmd)
	if err != nil {
		log.Fatalf("exec failed: exit=%d stdout=%q stderr=%q err=%v", res.ExitCode, res.Stdout, res.Stderr, err)
	}

	log.Printf("[runner] exit=%d stdout=%q stderr=%q err=%v", res.ExitCode, res.Stdout, res.Stderr, err)
}
