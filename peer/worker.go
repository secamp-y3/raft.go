package peer

import (
	"fmt"
	"math/rand"
	"sync"
)

type Worker struct {
	name string
	node *Node

	mu sync.Mutex

	State WorkerState
}

type WorkerOption func(*Worker)

func NewWorker(name string) *Worker {
	w := new(Worker)
	w.name = name
	w.State = InitState(w)
	return w
}

func (w *Worker) Name() string {
	return w.name
}

func (w *Worker) Addr() string {
	return w.node.Addr()
}

func (w *Worker) LockMutex() {
	w.mu.Lock()
}

func (w *Worker) UnlockMutex() {
	w.mu.Unlock()
}

func (w *Worker) Rand() *rand.Rand {
	return w.node.Rand()
}

func (w *Worker) LinkNode(n *Node) {
	w.node = n
}

func (w *Worker) Connect(name, addr string) (err error) {
	w.LockMutex()
	defer w.UnlockMutex()
	err = w.node.Connect(name, addr)
	if err != nil {
		return err
	}
	var reply RequestConnectReply
	err = w.RemoteCall(name, "Worker.RequestConnect", RequestConnectArgs{w.name, w.node.Addr()}, &reply)
	if err != nil {
		return err
	} else if !reply.OK {
		return fmt.Errorf("Connection request denied: [%s] %s", name, addr)
	}
	// for n, a := range reply.Peers {
	// 	if !w.node.IsConnectedTo(n) {
	// 		err = w.node.Connect(n, a)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	return nil
}

func (w *Worker) Stop() {
    w.node.Shutdown()
    w.node = nil
}

func (w *Worker) RemoteCall(name, method string, args any, reply any) error {
	return w.node.call(name, method, args, reply)
}

func (w *Worker) ConnectedPeers() map[string]string {
	return w.node.ConnectedNodes()
}

type RequestConnectArgs struct {
	Name string
	Addr string
}

type RequestConnectReply struct {
	OK bool
	Peers    map[string]string
}

func (w *Worker) RequestConnect(args RequestConnectArgs, reply *RequestConnectReply) error {
	reply.OK = false
	reply.Peers = make(map[string]string)
	w.LockMutex()
	defer w.UnlockMutex()
	err := w.node.Connect(args.Name, args.Addr)
	if err != nil {
		return err
	}
	reply.OK = true
	for name, addr := range w.node.ConnectedNodes() {
        reply.Peers[name] = addr
	}
	return nil
}

type RequestConnectedPeersArgs struct {}

type RequestConnectedPeersReply struct {
    Peers map[string]string
}

func (w *Worker) RequestConnectedPeers(args RequestConnectedPeersArgs, reply *RequestConnectedPeersReply) error {
    w.LockMutex()
    defer w.UnlockMutex()
    reply.Peers = make(map[string]string)
    for k, v := range w.ConnectedPeers() {
        reply.Peers[k] = v
    }
    return nil
}
