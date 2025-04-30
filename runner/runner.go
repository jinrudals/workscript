package runner

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Runner struct {
	Nodes map[string]*Node
}

// NewRunnerFromMap builds a Runner DAG from a parsed JSON map.
func NewRunnerFromMap(data map[string]map[string]interface{}) *Runner {
	r := &Runner{Nodes: make(map[string]*Node)}

	// 1. Create all nodes
	for name, nodeData := range data {
		run := extractStringMap(nodeData["run"])
		post := extractStringMap(nodeData["post"])
		r.Nodes[name] = &Node{Name: name, Command: run, Post: post}
	}

	// 2. Link dependencies
	for name, nodeData := range data {
		node := r.Nodes[name]

		if beforeList, ok := nodeData["before"].([]interface{}); ok {
			for _, b := range beforeList {
				child := r.Nodes[b.(string)]
				node.Children = append(node.Children, child)
				child.Parents = append(child.Parents, node)
				child.InDegree = max(child.InDegree, node.InDegree+1)
			}
		}

		if afterList, ok := nodeData["after"].([]interface{}); ok {
			for _, a := range afterList {
				parent := r.Nodes[a.(string)]
				node.Parents = append(node.Parents, parent)
				parent.Children = append(parent.Children, node)
				node.InDegree = max(node.InDegree, parent.InDegree+1)
			}
		}
	}
	return r
}

// Launch executes the DAG with concurrency and mode: "all", "post", or "command"
func (r *Runner) Launch(maxWorkers int, mode string) {

	totalNodes := len(r.Nodes)
	completedCount := 0
	statuses := map[string]string{}
	for name := range r.Nodes {
		statuses[name] = "pending"
	}

	// Create ordered name list for status printing
	var nodeOrder []string
	for name := range r.Nodes {
		nodeOrder = append(nodeOrder, name)
	}
	sort.Strings(nodeOrder)

	printStatusTable(statuses, nodeOrder, true)

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxWorkers)

	ready := make(chan *Node, totalNodes)

	// Seed ready queue with zero in-degree nodes
	for _, node := range r.Nodes {
		if node.InDegree == 0 {
			ready <- node
		}
	}

	for completedCount < totalNodes {
		select {
		case node := <-ready:
			if node.Executed {
				continue
			}

			// Prepare command info
			name, cmd, dir, postCmd, postDir := node.PrepareCommand()
			if mode == "post" {
				cmd = ""
			} else if mode == "command" {
				postCmd = ""
			}

			statuses[name] = "running"
			printStatusTable(statuses, nodeOrder, false)

			wg.Add(1)
			sem <- struct{}{}

			go func(n *Node, cmd, dir, postCmd, postDir string) {
				defer wg.Done()
				defer func() { <-sem }()

				info := CommandInfo{
					Name:           n.Name,
					Command:        cmd,
					CommandDir:     dir,
					PostCommand:    postCmd,
					PostCommandDir: postDir,
				}

				ExecuteCommand(info)

				mu.Lock()
				defer mu.Unlock()
				n.Executed = true
				statuses[n.Name] = "done"
				completedCount++
				printStatusTable(statuses, nodeOrder, false)

				for _, child := range n.Children {
					allParentsDone := true
					for _, p := range child.Parents {
						if !p.Executed {
							allParentsDone = false
							break
						}
					}
					if allParentsDone && !child.Executed {
						ready <- child
					}
				}
			}(node, cmd, dir, postCmd, postDir)

		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	wg.Wait()
	fmt.Println("All DAG stages executed.")
}

// String prints each node’s in-degree
func (r *Runner) String() string {
	var out strings.Builder
	nodes := []*Node{}
	for _, n := range r.Nodes {
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].InDegree < nodes[j].InDegree
	})
	for _, n := range nodes {
		fmt.Fprintf(&out, "%s: (in_degree=%d)\n", n.Name, n.InDegree)
	}
	return out.String()
}

func extractStringMap(raw interface{}) map[string]string {
	out := map[string]string{}
	if m, ok := raw.(map[string]interface{}); ok {
		for k, v := range m {
			out[k] = fmt.Sprintf("%v", v)
		}
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func printStatusTable(statuses map[string]string, order []string, firstTime bool) {
	if !firstTime {
		fmt.Printf("\033[%dA", len(order)) // move cursor up
	}
	for _, name := range order {
		fmt.Printf("\033[K%s: %s\n", name, statuses[name]) // clear line + print
	}
}
