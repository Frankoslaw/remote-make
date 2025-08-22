package domain

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/looplab/fsm"
)

type Node struct {
	ID   uuid.UUID
	Name string
	FSM  *fsm.FSM
}

func NewNode(name string) *Node {
	n := &Node{
		Name: name,
	}

	n.FSM = fsm.NewFSM("new", fsm.Events{
		// Public events the caller will use:
		{Name: "spawn", Src: []string{"new", "terminated", "failed"}, Dst: "spawning"},
		{Name: "terminate", Src: []string{"running", "spawning", "interrupted"}, Dst: "terminating"},
		{Name: "interrupt", Src: []string{"running", "spawning"}, Dst: "interrupted"},
		{Name: "fail", Src: []string{"spawning", "running", "terminating"}, Dst: "failed"},

		// Internal/transient events to reflect full lifecycle
		{Name: "start", Src: []string{"spawning"}, Dst: "running"},
		{Name: "finalize", Src: []string{"terminating"}, Dst: "terminated"},
	}, nil)
	fmt.Println(fsm.VisualizeForMermaidWithGraphType(n.FSM, fsm.FlowChart))

	return n
}

type NodeConn struct {
	PublicIP  string
	PrivateIP string
	Hostname  string
}
