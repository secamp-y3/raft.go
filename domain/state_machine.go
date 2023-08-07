package domain

import (
	"fmt"

	"github.com/secamp-y3/raft.go/server"
)

type Log string

type StateMachine struct {
	Node *server.Node
	Log  []Log
}

type AppendLogsArgs struct {
	Entries []Log
}

type AppendLogsReply int

type AppendEntriesArgs struct {
	Log []Log
}

type AppendEntriesReply int

func (s *StateMachine) AppendLogs(input AppendLogsArgs, reply *AppendLogsReply) error {
	s.Log = append(s.Log, input.Entries...)
	channel := s.Node.Channels()
	for _, c := range channel {
		var appendEntriesReply *AppendEntriesReply
		c.Call("StateMachine.AppendEntries", AppendEntriesArgs{Log: input.Entries}, appendEntriesReply)
	}
	fmt.Printf("Log: %v\n", s.Log)
	return nil
}

func (s *StateMachine) AppendEntries(input AppendEntriesArgs, reply *AppendEntriesReply) error {
	s.Log = append(s.Log, input.Log...)
	fmt.Printf("Log: %v\n", s.Log)
	return nil
}
