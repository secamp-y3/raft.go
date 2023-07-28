package peer

import (
	"math/rand"
	"sync"
	"time"
)

type Worker struct {
	state WorkerState
	peers map[string]string

	mu  sync.Mutex
	rnd *rand.Rand

	callChan    chan<- *Call
	connectChan chan<- *Connect
}

type WorkerOption func(*Worker)

func NewWorker(callChan chan<- *Call, connectChan chan<- *Connect) *Worker {
	w := new(Worker)

	w.rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	w.state = InitState(w)
	w.peers = make(map[string]string)
	w.callChan = callChan
	w.connectChan = connectChan

	return w
}

func (w *Worker) RegisterPeer(name string, addr string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.peers[name] = addr
}

func (w *Worker) UnregisterPeer(name string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.peers, name)
}

type Call struct {
	Name    string
	Method  string
	Args    any
	Reply   any
	ErrChan chan<- error
}

func (w *Worker) RemoteCall(name, method string, args any, reply any) error {
	errChan := make(chan error)
	call := &Call{name, method, args, reply, errChan}
	w.callChan <- call

	err := <-errChan
	return err
}

type RequestConnectArgs struct {
	Name string
	Addr string
}

type RequestConnectReply struct {
	OK bool
}

type Connect struct {
	Name    string
	Addr    string
	ErrChan chan<- error
}

func (w *Worker) RequestConnect(args RequestConnectArgs, reply *RequestConnectReply) error {
	reply.OK = false

	errChan := make(chan error)
	w.connectChan <- &Connect{args.Name, args.Addr, errChan}
	err := <-errChan
	if err != nil {
		return err
	}

	reply.OK = true
	return nil
}

type RequestConnectedPeersArgs struct{}

type RequestConnectedPeersReply struct {
	Peers map[string]string
}

func (w *Worker) RequestConnectedPeers(args *RequestConnectedPeersArgs, reply *RequestConnectedPeersReply) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	reply.Peers = make(map[string]string)
	for name, addr := range w.peers {
		reply.Peers[name] = addr
	}

	return nil
}
