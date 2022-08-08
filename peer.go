package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

    "sc.y3/peer"
    "sc.y3/dispatcher"
)

func main() {
    host, err := os.Hostname()
    if err != nil {
        log.Fatal(err)
    }

	nameFlag := flag.String("name", host, "Worker name")
	portFlag := flag.Int("port", 30000, "Port number")
    delayFlag := flag.Int("delay", 0, "Communication delay")
    dispatcherFlag := flag.String("dispatcher", "", "Dispathcer address")
	flag.Parse()

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, *portFlag))
    if err != nil {
        log.Fatal(err)
    }

    log.Println(addr.String())

    node := peer.NewNode(addr.String(), peer.Delay(*delayFlag))
    dispatcher, err := dispatcher.FindDispatcher(*dispatcherFlag)
    if err != nil {
        log.Fatal(err)
    }

    worker := peer.NewWorker(*nameFlag)
    if node.LinkWorker(worker) != nil {
        log.Fatal(err)
    }

    peers, err := dispatcher.GetConnectedPeers(worker)
    if err != nil {
        log.Fatal(err)
    }
    for name, addr := range peers {
        if name != worker.Name() {
            worker.Connect(name, addr)
        }
    }

    for {}
}
