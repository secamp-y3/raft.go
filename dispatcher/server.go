package dispatcher

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
)

type Server struct {
	peers map[string]string
	mu    sync.Mutex
}

func StartDispatcher(addr string) error {
	s := new(Server)
	s.peers = make(map[string]string)
	server := rpc.NewServer()
	server.RegisterName("Dispatcher", s)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go server.ServeConn(conn)
	}
}

type RequestConnectArgs struct {
	Name string
	Addr string
}

type RequestConnectReply struct {
	Peers map[string]string
}

func (s *Server) RequestConnect(args RequestConnectArgs, reply *RequestConnectReply) error {
	log.Printf("Connection request from %s:%s\n", args.Name, args.Addr)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.peers[args.Name] = args.Addr
	reply.Peers = make(map[string]string)
	for k, v := range s.peers {
		reply.Peers[k] = v
	}
	return nil
}

type RequestAddrArgs struct {
	Name string
}

type RequestAddrReply struct {
	Addr string
}

func (s *Server) RequestAddr(args RequestAddrArgs, reply *RequestAddrReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	addr, ok := s.peers[args.Name]
	if !ok {
		return fmt.Errorf("Not found: %s", args.Name)
	}
	reply.Addr = addr
	return nil
}
