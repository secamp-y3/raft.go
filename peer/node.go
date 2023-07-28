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

type connectedNode struct {
	addr string
	conn *rpc.Client
}

type Node struct {
	name string
	addr string

	listener net.Listener
	server   *rpc.Server
	peers    map[string]*connectedNode

	mu sync.Mutex
	wg sync.WaitGroup

	rnd *rand.Rand

	worker      *Worker
	callChan    <-chan *Call
	connectChan <-chan *Connect
	quit        chan interface{}

	delay int
}

type NodeOption func(*Node)

func Delay(t int) NodeOption {
	return func(n *Node) {
		n.delay = t
	}
}

func NewNode(name, addr string, callChan <-chan *Call, connectChan <-chan *Connect, worker *Worker, options ...NodeOption) (n *Node, err error) {
	n = new(Node)

	n.name = name
	n.addr = addr
	n.peers = make(map[string]*connectedNode)
	n.delay = 0
	n.quit = make(chan interface{})
	n.rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	n.worker = worker
	n.callChan = callChan
	n.connectChan = connectChan
	n.server = rpc.NewServer()
	err = n.server.RegisterName("Worker", n.worker)
	if err != nil {
		return nil, err
	}
	n.listener, err = net.Listen("tcp", n.addr)
	if err != nil {
		return nil, err
	}

	for _, opt := range options {
		opt(n)
	}

	return
}

func (n *Node) Name() string {
	return n.name
}

func (n *Node) Addr() string {
	return n.addr
}

func (n *Node) isConnectedTo(name string) bool {
	_, ok := n.peers[name]
	return ok
}

func (n *Node) Connect(name, addr string) error {
	if !n.isConnectedTo(name) {
		log.Printf("Connect to %s:%s", name, addr)
		n.mu.Lock()
		defer n.mu.Unlock()

		c, err := rpc.Dial("tcp", addr)
		if err != nil {
			return err
		}

		n.peers[name] = &connectedNode{addr, c}
		n.worker.RegisterPeer(name, addr)
	}

	var reply RequestConnectReply
	err := n.call(name, "Worker.RequestConnect", RequestConnectArgs{n.name, n.addr}, &reply)
	if err != nil {
		return err
	} else if !reply.OK {
		return fmt.Errorf("Connection request denied: [%s] %s", name, addr)
	}

	log.Printf("Connected to %s:%s", name, addr)
	return nil
}

func (n *Node) Disconnect(name string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if peer, ok := n.peers[name]; !ok || peer == nil {
		return nil
	}

	err := n.peers[name].conn.Close()
	if err != nil {
		return err
	}

	delete(n.peers, name)
	n.worker.UnregisterPeer(name)
	return nil
}

func (n *Node) call(name, method string, args any, reply any) error {
	n.mu.Lock()
	peer := n.peers[name]
	n.mu.Unlock()
	if peer == nil {
		return fmt.Errorf("No such peer: %s", name)
	}

	if n.delay > 0 {
		delay := int(n.rnd.ExpFloat64() / float64(n.delay))
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

	return peer.conn.Call(method, args, reply)
}

func (n *Node) Run() error {
	// RPCサーバーを起動
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

	// workerからのchanによるリクエストを処理
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		for {
			select {
			case <-n.quit:
				return

			case call := <-n.callChan:
				err := n.call(call.Name, call.Method, call.Args, call.Reply)
				call.ErrChan <- err

			case connect := <-n.connectChan:
				err := n.Connect(connect.Name, connect.Addr)
				connect.ErrChan <- err
			}
		}
	}()

	return nil
}

func (n *Node) Shutdown() {
	close(n.quit)
	n.listener.Close()

	for name := range n.peers {
		_ = n.Disconnect(name)
	}
	n.worker = nil
	n.wg.Wait()
}
