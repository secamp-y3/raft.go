package main

import (
	"log"
	"net/rpc"
	"os"

	"github.com/secamp-y3/raft.go/domain"
	"github.com/secamp-y3/raft.go/server"
	"github.com/spf13/pflag"
)

func getEnvOr(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func main() {
	name := pflag.StringP("name", "n", getEnvOr("NAME", "node"), "Name of this node")
	host := pflag.StringP("host", "h", getEnvOr("HOST", "localhost"), "Host name")
	port := pflag.StringP("port", "p", getEnvOr("PORT", "8080"), "Port to listen")

	initiatorAddr := pflag.StringP("server", "s", "", "Server address to join P2P network")

	meanDelay := pflag.Float64P("delay", "d", 0, "Mean delay of communication channel (Unit: ms)")
	lossRate := pflag.Float64P("loss", "l", 0, "Loss rate of communication channel")
	seed := pflag.Int64P("seed", "", 0, "Random seed")

	pflag.Parse()

	node, err := server.NewNode(*name, *host, *port, server.MeanDelay(*meanDelay), server.LossRate(*lossRate), server.Seed(*seed))
	if err != nil {
		log.Fatal(err)
	}

	heartbeatWatch := make(chan int, 100)
	stateMachine := domain.StateMachine{Node: node, Log: []domain.Log{}, HeartbeatWatch: heartbeatWatch, Term: 0, Role: "follower"}

	svr := rpc.NewServer()
	svr.RegisterName("Monitor", &domain.Monitor{Node: node})
	svr.RegisterName("StateMachine", &stateMachine)

	shutdown := node.Serve(svr)
	defer shutdown()

	if *initiatorAddr != "" {
		if err := node.EstablishConnection(server.Addr(*initiatorAddr)); err != nil {
			log.Fatal(err)
		}
	}

	for {
		switch stateMachine.Role {
		case "follower":
			stateMachine.ExecFollower(heartbeatWatch)
		case "candidate":
			stateMachine.ExecCandidate()
		case "leader":
			stateMachine.ExecLeader()
		}
	}
}
