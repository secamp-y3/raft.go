package heartbeat

import (
	"fmt"
	"time"

	"github.com/secamp-y3/raft.go/domain"
	"github.com/secamp-y3/raft.go/server"
)

type HeartBeat struct {
	Node *server.Node
}

func (h *HeartBeat) HeartBeat(stateMachine *domain.StateMachine) {
	fmt.Printf("HeartBeat Send: Term: %d, Role: %s, Leader: %s\n", stateMachine.Term, stateMachine.Role, stateMachine.Leader)
	for {
		if stateMachine.Role != "leader" {
			break
		}
		fmt.Printf("Channel: %v\n", h.Node.Channels())
		channel := h.Node.Channels()
		for _, ch := range channel {
			appendEntriesReply := &domain.AppendEntriesReply{}
			err := ch.Call("StateMachine.AppendEntries", domain.AppendEntriesArgs{Term: stateMachine.Term}, appendEntriesReply)
			if err != nil {
				fmt.Printf("Failed to send heartbeat: %v\n", err)
			}
		}
		time.Sleep(1 * time.Second)
	}
}
