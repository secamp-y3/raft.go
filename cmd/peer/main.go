package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"sc.y3/dispatcher"
	"sc.y3/peer"
)

func main() {
	host, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	nameFlag := flag.String("name", host, "Worker name")
	portFlag := flag.Int("port", 30000, "Port number")
	delayFlag := flag.Int("delay", 0, "Communication delay")
	dispatcherFlag := flag.String("dispatcher", "", "Dispatcher address")
	flag.Parse()

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, *portFlag))
	if err != nil {
		log.Fatal(err)
	}

	log.Println(addr.String())

	callChan := make(chan *peer.Call)
	connectChan := make(chan *peer.Connect)
	worker := peer.NewWorker(callChan, connectChan)
	node, err := peer.NewNode(*nameFlag, addr.String(), callChan, connectChan, worker, peer.Delay(*delayFlag))
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		err := node.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	dispatcher, err := dispatcher.FindDispatcher(*dispatcherFlag)
	if err != nil {
		log.Fatal(err)
	}

	peers, err := dispatcher.GetConnectedPeers(node)
	if err != nil {
		log.Fatal(err)
	}
	for name, addr := range peers {
		if name != node.Name() {
			_ = node.Connect(name, addr)
			_ = node.ConnectBack(name, addr)
		}
	}

	select {}
}
