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

	meanDelay := pflag.Float64P("delay", "d", 0, "Mean delay of communication channel")
	lossRate := pflag.Float64P("loss", "l", 0, "Loss rate of communication channel")
	seed := pflag.Int64P("seed", "", 0, "Random seed")

	pflag.Parse()

	node, err := server.NewNode(*name, *host, *port, server.MeanDelay(*meanDelay), server.LossRate(*lossRate), server.Seed(*seed))
	if err != nil {
		log.Fatal(err)
	}

	svr := rpc.NewServer()
	svr.RegisterName("Monitor", &domain.Monitor{Node: node})

	shutdown := node.Serve(svr)
	defer shutdown()

	if *initiatorAddr != "" {
		if err := node.EstablishConnection(server.Addr(*initiatorAddr)); err != nil {
			log.Fatal(err)
		}
	}
}
