package domain

type Command struct {
	Args  []string `json:"args"`
	Stdin []byte   `json:"stdin,omitempty"`
}

type ExecResult struct {
	ExitCode int    `json:"exit_code"`
	Stdout   []byte `json:"stdout"`
	Stderr   []byte `json:"stderr"`
	Err      string `json:"err,omitempty"`
}
