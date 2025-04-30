package runner

import (
	"fmt"
	"os"
)

var cwd, _ = os.Getwd()

type Node struct {
	Name     string
	Command  map[string]string
	Post     map[string]string
	InDegree int
	Parents  []*Node
	Children []*Node

	Executed bool
}

func (n *Node) PrepareCommand() (string, string, string, string, string) {
	cmdDir := cwd
	postDir := cwd
	cmd := ""
	postCmd := ""

	if val, ok := n.Command["directory"]; ok {
		cmdDir = val
	}
	if val, ok := n.Command["command"]; ok {
		cmd = val
	}
	if val, ok := n.Post["directory"]; ok {
		postDir = val
	}
	if val, ok := n.Post["command"]; ok {
		postCmd = val
	}

	return n.Name, cmd, cmdDir, postCmd, postDir
}

// String returns a debug-friendly string version of the node
func (n *Node) String() string {
	return fmt.Sprintf("Node(%s)", n.Name)
}
