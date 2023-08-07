package server

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/secamp-y3/raft.go/logger"
)

// Node manages network connections as a representation of a peer in P2P network
type Node struct {
	NodeInfo
	network *Cluster

	meanDelay float64
	lossRate  float64
	rng       *rand.Rand

	listener net.Listener
	logger   logger.Logger
}

type NodeOption func(*Node)

func MeanDelay(d float64) NodeOption {
	return func(n *Node) {
		n.meanDelay = d
	}
}

func LossRate(l float64) NodeOption {
	return func(n *Node) {
		n.lossRate = l
	}
}

func Seed(s int64) NodeOption {
	return func(n *Node) {
		if s > 0 {
			n.rng.Seed(s)
		}
	}
}

// NewNode returns a pointer to new node if the given host and port are valid
func NewNode(name, host, port string, opts ...NodeOption) (*Node, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	node := &Node{
		NodeInfo:  NodeInfo{name, Addr(addr.String())},
		network:   NewCluster(),
		listener:  listener,
		meanDelay: 0,
		lossRate:  0,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
		logger:    logger.Logger{Name: fmt.Sprintf("Node %s@%s", name, addr.String())},
	}
	for _, opt := range opts {
		opt(node)
	}

	node.logger.Info("Initialized with {delay: %f, loss: %f}", node.meanDelay, node.lossRate)

	return node, nil
}

// Self returns the name and endpoint address of the node
func (n *Node) Self() NodeInfo {
	return n.NodeInfo
}

// Network gives a pointer to cluster manager
//
// Use `(*Node).Channels()` to send RPC to other nodes in the cluster.
func (n *Node) Network() *Cluster {
	return n.network
}

// Random gives a pointer to the pseudo ranndom generator owned by the node
func (n *Node) Random() *rand.Rand {
	return n.rng
}

// Channels returns a list of communication channels to cluster members
func (n *Node) Channels() map[string]Channel {
	members := n.Network().Members()
	channels := map[string]Channel{}
	for _, member := range members {
		switch {
		case n.lossRate > 0 && n.rng.Float64() < n.lossRate:
			channels[member.Name] = NewLostChannel(member.Endpoint)
		case n.meanDelay > 0:
			channels[member.Name] = NewDelayedChannel(member.Endpoint, n.rng.ExpFloat64()*n.meanDelay)
		default:
			channels[member.Name] = NewReliableChannel(member.Endpoint)
		}
	}
	return channels
}

// EstablishConnection connects this node to the other node with given address
func (n *Node) EstablishConnection(addr Addr) error {
	n.logger.Info("Trying to join network via %s", addr.String())
	if err := n.establishConnection(addr); err != nil {
		return err
	}
	n.logger.Info("Joined the network with %v", n.Network().Members())
	return nil
}

func (n *Node) establishConnection(addr Addr) error {
	if addr == n.Endpoint {
		n.logger.Error("Cannot request self connection")
		return nil
	}
	n.logger.Info("Request connection to %s", addr.String())
	ch := NewReliableChannel(addr)
	reply := RequestConnectionReply{}
	if err := ch.Call("__Node.RequestConnection", RequestConnectionArgs{From: n.NodeInfo}, &reply); err != nil {
		return err
	}
	if !reply.Accepted {
		n.logger.Error("Connection to %s is refused", addr.String())
		return nil
	}
	if !n.Network().Append(NodeInfo{Name: reply.Name, Endpoint: addr}).Success() {
		n.logger.Error("Node % s was somehow already connected", reply.Name)
		return nil
	}
	n.logger.Info("Connection to %s is accepted", NodeInfo{Name: reply.Name, Endpoint: addr})
	wg := sync.WaitGroup{}
	for _, node := range reply.Members {
		if node != n.Self() && n.Network().Append(node).Success() {
			wg.Add(1)
			go func(ni NodeInfo) {
				n.establishConnection(ni.Endpoint)
				wg.Done()
			}(node)
		}
	}
	wg.Wait()
	return nil
}

// ShutdownFunc should be called before terminating the program
type ShutdownFunc func()

// Serve start the given server on this node
func (n *Node) Serve(server *rpc.Server) ShutdownFunc {
	server.RegisterName("__Node", n)

	wg := sync.WaitGroup{}
	tc := make(chan interface{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	go n.mainLoop(server, ctx, &wg, tc)

	n.logger.Info("Start server")

	sgnl := make(chan os.Signal, 1)
	signal.Ignore()
	signal.Notify(sgnl, syscall.SIGINT)

	return func() {
		s := <-sgnl
		switch s {
		case syscall.SIGINT:
			n.logger.Info("Signal interrupt received.")
			cancel()
			n.listener.Close()
			n.logger.Info("Shutting down...")
			<-tc
			wg.Wait()
		default:
			n.logger.Fatal("Unexpected signal")
		}
		n.logger.Info("Shutdown")
	}
}

func (n *Node) mainLoop(server *rpc.Server, ctx context.Context, wg *sync.WaitGroup, tc chan interface{}) {
	defer close(tc)
	for {
		conn, err := n.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			if errors.Is(err, net.ErrClosed) {
				n.logger.Error(err.Error())
				<-ctx.Done()
				return
			}
		}
		n.logger.Info("Connection from %s", conn.RemoteAddr().String())

		wg.Add(1)
		go func() {
			defer func() {
				conn.Close()
				wg.Done()
			}()

			worker := make(chan struct{}, 1)
			go func() {
				defer close(worker)
				server.ServeConn(conn)
			}()

			select {
			case <-ctx.Done():
			case <-worker:
			}
		}()
	}
}

type RequestConnectionArgs struct {
	From NodeInfo
}

type RequestConnectionReply struct {
	Accepted bool
	Name     string
	Members  []NodeInfo
}

func (n *Node) RequestConnection(args RequestConnectionArgs, reply *RequestConnectionReply) error {
	n.logger.Info("Received: connection request from %s", args.From.String())
	reply.Accepted = n.Network().Append(args.From).Success()
	if !reply.Accepted {
		n.logger.Error("Connection request from %s was rejected.", args.From.String())
		return nil
	}
	n.logger.Info("Connection request from %s is accepted", args.From.String())
	reply.Name = n.Self().Name
	reply.Members = n.Network().Members()
	n.logger.Info("Replying with %v", reply.Members)
	return nil
}
