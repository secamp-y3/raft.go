package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"sc.y3/dispatcher"
)

func main() {
	portFlag := flag.Int("port", 8080, "Port number")
	flag.Parse()

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", *portFlag))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Dispatcher: %s", addr.String())
	if err := dispatcher.StartDispatcher(addr.String()); err != nil {
		log.Fatal(err)
	}
}
