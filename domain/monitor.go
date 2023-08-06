package domain

import (
	"fmt"
	"strings"

	"github.com/secamp-y3/raft.go/server"
)

type Monitor struct {
	Node *server.Node
}

func (m *Monitor) FetchState(_ FetchStateArgs, reply *FetchStateReply) error {
	if m.Node == nil {
		return fmt.Errorf("Monitor is not initialized properly.")
	}

	// node state
	reply.NodeInfo = m.Node.Self()
	reply.Members = m.Node.Network().Members()

	return nil
}

type FetchStateArgs struct{}

type FetchStateReply struct {
	NodeInfo server.NodeInfo   `json:"node_info"`
	Members  []server.NodeInfo `json:"members"`
}

func (r *FetchStateReply) String() string {
	sb := strings.Builder{}
	sb.WriteString(r.NodeInfo.String())
	sb.WriteString(": [")
	for i, node := range r.Members {
		sb.WriteString(node.String())
		if i < len(r.Members)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]")
    return sb.String()
	// return fmt.Sprintf("%s: %v", r.NodeInfo, r.Members)
}
