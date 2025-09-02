package domain

import "io"

type ExecProviderID string
type ExecHandleID string

type AttachRequest struct {
	NodeID         NodeID
	ExecProviderID ExecProviderID

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}
type AttachResult struct {
	ExitCode <-chan int
	Close    func() error
	Wait     func() error
}

type ExecRequest struct {
	NodeID         NodeID
	ExecProviderID ExecProviderID

	Command    []string
	Env        map[string]string
	WorkingDir string
}
type ExecResult struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
}

type CopyToRequest struct {
	NodeID         NodeID
	ExecProviderID ExecProviderID

	Src string
	Dst string
}
type CopyFromRequest struct {
	NodeID         NodeID
	ExecProviderID ExecProviderID

	Src string
	Dst string
}
