package peer

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type ConnectedNode struct {
	addr string
	conn *rpc.Client
}

type Node struct {
	addr  string

	listener net.Listener
	server   *rpc.Server
	peers    map[string]*ConnectedNode

	mu sync.Mutex
	wg sync.WaitGroup

	rnd *rand.Rand

	worker *Worker
	quit   chan interface{}

	delay int
	verbose bool
}

type NodeOption func(*Node)

func Delay(t int) NodeOption {
	return func(n *Node) {
		n.delay = t
	}
}

func NewNode(addr string, options ...NodeOption) *Node {
	n := new(Node)
	n.addr = addr
	n.peers = make(map[string]*ConnectedNode)
	n.delay = 0
	n.quit = make(chan interface{})
	n.rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, opt := range options {
		opt(n)
	}
	return n
}

func (n *Node) Addr() string {
	return n.addr
}

func (n *Node) Rand() *rand.Rand {
	return n.rnd
}

func (n *Node) Connect(name, addr string) error {
	if n.IsConnectedTo(name) {
		return fmt.Errorf("Peer '%s' is already reserved", name)
	}
    log.Printf("Connect to %s:%s", name, addr)
	n.mu.Lock()
	defer n.mu.Unlock()
	c, err := rpc.Dial("tcp", addr)
	if err != nil {
		return err
	}
	n.peers[name] = &ConnectedNode{addr, c}
	return nil
}

func (n *Node) Shutdown() {
	close(n.quit)
	n.listener.Close()
	for name, _ := range n.ConnectedNodes() {
	    n.Disconnect(name)
    }
    n.worker = nil
	n.wg.Wait()
}

func (n *Node) Disconnect(name string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.peers[name] != nil {
		err := n.peers[name].conn.Close()
		delete(n.peers, name)
		return err
	}
	return nil
}

func (n *Node) LinkWorker(w *Worker) error {
	n.mu.Lock()
	n.worker = w
	n.worker.LinkNode(n)

	n.server = rpc.NewServer()
	n.server.RegisterName("Worker", n.worker)

	var err error
	n.listener, err = net.Listen("tcp", n.addr)
	if err != nil {
		return err
	}
	n.mu.Unlock()

	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		for {
			conn, err := n.listener.Accept()
			if err != nil {
				select {
				case <-n.quit:
					return
				default:
					log.Fatal(err)
				}
			}
			n.wg.Add(1)
			go func() {
				n.server.ServeConn(conn)
				n.wg.Done()
			}()
		}
	}()

	return nil
}

func (n *Node) ConnectedNodes() map[string]string {
	n.mu.Lock()
	defer n.mu.Unlock()
	list := make(map[string]string, len(n.peers))
	for key, value := range n.peers {
		list[key] = value.addr
	}
	return list
}

func (n *Node) ConnectedNodeNames() []string {
	n.mu.Lock()
	defer n.mu.Unlock()
	ret := make([]string, len(n.peers))
	for key := range n.peers {
		ret = append(ret, key)
	}
	return ret
}

func (n *Node) IsConnectedTo(name string) bool {
    _, ok := n.peers[name]
	return ok
}

func (n *Node) call(name, method string, args any, reply any) error {
	n.mu.Lock()
	peer := n.peers[name]
	n.mu.Unlock()
	if peer == nil {
		return fmt.Errorf("No such peer: %s", name)
	}
	if n.delay > 0 {
		delay := int(n.Rand().ExpFloat64() / float64(n.delay))
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	return peer.conn.Call(method, args, reply)
}
