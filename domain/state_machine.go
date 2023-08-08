package domain

import (
	"fmt"
	"log"
	"math/rand"
	"time"

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
	if input.Term < s.Term {
		fmt.Printf("RequestVote failed. InputTerm: %d, Term: %d\n", input.Term, s.Term)
		return nil
	}
	s.Term = input.Term
	s.Leader = input.Leader
	s.Role = "follower"
	reply.VoteGranted = true
	fmt.Printf("Role began follower, Term: %d, Role: %s, Leader: %s\n", s.Term, s.Role, s.Leader)
	return nil
}

func (s *StateMachine) HeartBeat() {
	fmt.Println("HeartBeat")
	channel := s.Node.Channels()
	fmt.Printf("Channel: %v\n", channel)
	for k, ch := range channel {
		appendEntriesReply := &AppendEntriesReply{}
		println(1)
		err := ch.Call("StateMachine.AppendEntries", AppendEntriesArgs{Term: s.Term}, appendEntriesReply)
		println(2)
		if err != nil {
			fmt.Printf("Failed to send heartbeat: %v\n", err)
			s.Node.Network().Remove(k)
		}
	}
	time.Sleep(1 * time.Second)
}

func (s *StateMachine) ExecLeader() {
	s.HeartBeat()
}

func (s *StateMachine) ExecFollower(heartbeatWatch chan int) {
	seed := time.Now().UnixNano()
	r := rand.New(rand.NewSource(seed))
	val := r.Intn(1000) + 2000
	select {
	case v := <-heartbeatWatch:
		if v == 1 {
			log.Println("Heartbeat is working")
		} else {
			log.Println("Heartbeat is not working(heartbeatWatch)")
		}
	case <-time.After(time.Duration(val) * time.Millisecond):
		if s.Role == "leader" {
			return
		}
		log.Println("Timeout")
		log.Println("Heartbeat is not working")
		s.Term++
		s.Role = "candidate"
		fmt.Printf("Role began candidate: Role: %s, Leader: %s, Term: %d\n", s.Role, s.Leader, s.Term)
	}
}

func (s *StateMachine) ExecCandidate() {
	channels := s.Node.Channels()
	voteGrantedCnt := 0
	for k, c := range channels {
		requestVoteReply := RequestVoteReply{}
		err := c.Call("StateMachine.RequestVote", RequestVoteArgs{Term: s.Term, Leader: s.Node.Name}, &requestVoteReply)
		if err != nil {
			fmt.Printf("RequestVote Error: %v\n", err)
			s.Node.Network().Remove(k)

			fmt.Printf("key: %v, Channel: %v\n", k, s.Node.Channels())
			continue
		}
		if requestVoteReply.VoteGranted {
			voteGrantedCnt++
		}
	}
	if voteGrantedCnt > len(channels)/2 {
		s.Role = "leader"
		s.Leader = s.Node.Name
	}
	fmt.Printf("Role began leader: Role: %s, Leader: %s, Term: %d\n", s.Role, s.Leader, s.Term)
}
