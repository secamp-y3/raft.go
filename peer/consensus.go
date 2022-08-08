package peer

import (
	"fmt"
	"log"
)

type WorkerState struct {
	Value int
}

func InitState(w *Worker) WorkerState {
	return WorkerState{0}
}

func (s *WorkerState) String() string {
	return fmt.Sprintf("Value: %d", s.Value)
}

type RequestStateArgs struct{}

type RequestStateReply struct {
	State WorkerState
}

func (w *Worker) RequestState(args RequestStateArgs, reply *RequestStateReply) error {
	w.LockMutex()
	reply.State = w.State
	w.UnlockMutex()

	log.Println(reply.State.String())

	return nil
}
