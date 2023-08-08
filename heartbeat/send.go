package heartbeat

import (
	"log"
	"time"

	"github.com/secamp-y3/raft.go/domain"
	"github.com/secamp-y3/raft.go/server"
)

type HeartBeat struct {
	Node *server.Node
}

func (h *HeartBeat) HeartBeat() {
	time.Sleep(5 * time.Second)
	channel := h.Node.Channels()
	for {
		// fmt.Printf("Channel: %v\n", channel)
		for _, ch := range channel {
			appendEntriesReply := &domain.AppendEntriesReply{}
			err := ch.Call("StateMachine.AppendEntries", domain.AppendEntriesArgs{}, appendEntriesReply)
			if err != nil {
				log.Fatalf("Failed to send heartbeat: %v", err)
			}
		}
		time.Sleep(1 * time.Second)
	}
}
