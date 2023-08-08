package domain

import (
	"fmt"

	"github.com/secamp-y3/raft.go/server"
)

type Log string

type StateMachine struct {
	Node           *server.Node
	Log            []Log
	HeartbeatWatch chan int
	Term           int
	Leader         string
	Role           string
}

type AppendLogsArgs struct {
	Entries []Log
}

type AppendLogsReply int

type AppendEntriesArgs struct {
	Log  []Log
	Term int
}

type AppendEntriesReply struct{}

type RequestVoteArgs struct {
	Term   int
	Leader string
}

type RequestVoteReply struct {
	VoteGranted bool
}

func (s *StateMachine) AppendLogs(input AppendLogsArgs, reply *AppendLogsReply) error {
	s.Log = append(s.Log, input.Entries...)
	channel := s.Node.Channels()
	for _, c := range channel {
		appendEntriesReply := &AppendEntriesReply{}
		c.Call("StateMachine.AppendEntries", AppendEntriesArgs{Log: input.Entries}, appendEntriesReply)
	}
	fmt.Printf("Log: %v\n", s.Log)
	return nil
}

func (s *StateMachine) AppendEntries(input AppendEntriesArgs, reply *AppendEntriesReply) error {
	if input.Term < s.Term {
		return nil
	}
	s.Term = input.Term
	s.Role = "follower"
	s.HeartbeatWatch <- 1
	s.Log = append(s.Log, input.Log...)
	fmt.Printf("Log: %v\n", s.Log)
	return nil
}

func (s *StateMachine) RequestVote(input RequestVoteArgs, reply *RequestVoteReply) error {
	fmt.Println("RequestVote Start")
	if input.Term <= s.Term {
		fmt.Printf("RequestVote failed. InputTerm: %d, Term: %d\n", input.Term, s.Term)
		return nil
	}
	s.Term = input.Term
	s.Leader = input.Leader
	s.Role = "follower"
	reply.VoteGranted = true
	fmt.Printf("RequestVote succeed Term: %d, Role: %s, Leader: %s\n", s.Term, s.Role, s.Leader)
	return nil
}
