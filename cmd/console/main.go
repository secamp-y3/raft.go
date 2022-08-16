package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"

	"sc.y3/dispatcher"
	"sc.y3/peer"
)

var (
	disp *dispatcher.Client
)

func main() {
	dispatcherFlag := flag.String("dispatcher", "localhost:8080", "Dispatcher address")
	flag.Parse()

	if *dispatcherFlag == "localhost:8080" && os.Getenv("DISPATCHER") != "" {
		*dispatcherFlag = os.Getenv("DISPATCHER")
	}

	var err error
	disp, err = dispatcher.FindDispatcher(*dispatcherFlag)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("? ")
		scanner.Scan()
		input := scanner.Text()
		command := parse(input)
		result, err := command.Exec()
		if err != nil {
			fmt.Printf("> [ERROR] %s\n", err)
		} else {
			fmt.Println("> ", result)
		}
	}
}

func parse(raw string) Command {
	ret := Command{"", make([]string, 0)}
	for i, v := range strings.Split(raw, " ") {
		switch i {
		case 0:
			ret.operation = v
		default:
			ret.args = append(ret.args, v)
		}
	}
	return ret
}

func sendRPC(peer, method string, args any, reply any) error {
	addr, err := disp.GetAddr(peer)
	if err != nil {
		return err
	}
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		return err
	}

	return client.Call(method, args, reply)
}

type Command struct {
	operation string
	args      []string
}

func (c *Command) Exec() (string, error) {
	switch c.operation {
	case "state":
		return State(c.args[0])
	default:
		return "", fmt.Errorf("No such command")
	}
}

func State(name string) (string, error) {
	var reply peer.RequestStateReply
	err := sendRPC(name, "Worker.RequestState", peer.RequestStateArgs{}, &reply)
	if err != nil {
		return "", err
	}
	return reply.State.String(), nil
}
