package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"remoteMake/internal/infra"
	"remoteMake/internal/model"
	"remoteMake/internal/service"
	"remoteMake/internal/service/notify"
	"remoteMake/internal/service/posix"
)

func main() {
	uidGen := service.NewUIDService()
	_, _ = infra.ConnectDB("test.db")
	_, _ = infra.ConnectPodman("unix:///run/podman/podman.sock")

	taskManager := service.NewTaskManager(uidGen)
	taskManager.RegisterRunner("posix", posix.NewRunner(uidGen))

	notificationManager := service.NewNotificationManager()
	notificationManager.RegisterNotifier("discord", notify.NewDiscordNotifier("https://discord.com/api/webhooks/1399467506304815268/KP-p7Clnm_xD3eYTdSQrH98Vo_N_VgYB7654SYIN_27ZZp0XkFdh_hg3BbmkkpSWK1zA"))

	spec := model.Spec{
		Command: []string{"ls", "-lah"},
		Dir:     "/home/frankoslaw/Documents/programming/projects/remote-make",
	}

	task, err := taskManager.CreateTask("posix", spec)
	if err != nil {
		log.Fatalf("failed to create process: %v", err)
	}

	proc := task.Proc
	proc.OnExit(func(s model.State) {
		var stdoutBuf, stderrBuf bytes.Buffer

		_, _ = io.Copy(&stdoutBuf, proc.Stdout())
		_, _ = io.Copy(&stderrBuf, proc.Stderr())

		stdout := stdoutBuf.String()
		stderr := stderrBuf.String()

		if stdout == "" {
			stdout = "(no output)"
		}
		if stderr == "" {
			stderr = "(no errors)"
		}

		fmt.Printf(
			"Process ID: `%s`\nExit Code: `%d`\nError: `%v`\n\nStdout:\n%s\nStderr:\n%s\n",
			proc.ID(), s.ExitCode, s.Err, stdout, stderr,
		)
	})

	stateCh, err := proc.Start(context.Background())
	if err != nil {
		log.Fatalf("failed to start process: %v", err)
	}

	select {
	case state := <-stateCh:
		_ = state
	case <-time.After(5 * time.Second):
		_ = proc.Stop()
		log.Printf("Process %s timed out", proc.ID())
	}
}
