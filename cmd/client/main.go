package main

import (
	"bufio"
	"fmt"
	"net/rpc"
	"os"
	"strings"

	"github.com/secamp-y3/raft.go/domain"
	"github.com/spf13/pflag"
)

func sendRPC(dest, method string, args, reply any) error {
	client, err := rpc.Dial("tcp", dest)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Call(method, args, reply)
}

func FetchState(args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("Target server address is not given")
	}

	var reply domain.FetchStateReply
	if err := sendRPC(args[0], "Monitor.FetchState", domain.FetchStateArgs{}, &reply); err != nil {
		return "", err
	}

	return reply.String(), nil
}

func AppendLogs(args []string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("Target server address or log entry is not given")
	}

	var reply domain.AppendLogsReply
	logArgs := args[1:]
	var log []domain.Log = make([]domain.Log, len(logArgs))
	for i, v := range logArgs {
		log[i] = domain.Log(v)
	}
	if err := sendRPC(args[0], "StateMachine.AppendLogs", domain.AppendLogsArgs{Entries: log}, &reply); err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", reply), nil
}

// Handler represents callback functions for command
type Handler func(args []string) (string, error)

// HandlerMap stores command name with corresponsing handler
type HandlerMap map[string]Handler

// App manages a map from command to handler
type App struct {
	handlers HandlerMap
}

// Register registers the given command with the corresponding handler
func (a *App) Register(command string, handler Handler) {
	a.handlers[command] = handler
}

// Exec executes the command
func (a *App) Exec(c Command) (string, error) {
	if handler, ok := a.handlers[c.Name]; ok {
		return handler(c.Args)
	}
	return "", fmt.Errorf("Unknown command: %s", c.Name)
}

// Command is a pair of the command name and list of arguments
type Command struct {
	Name string
	Args []string
}

// ParseCommand creates a command from the input string
func ParseCommand(input string) Command {
	ret := Command{Name: "", Args: make([]string, 0)}
	for i, v := range strings.Split(input, " ") {
		switch i {
		case 0:
			ret.Name = v
		default:
			ret.Args = append(ret.Args, v)
		}
	}
	return ret
}

func main() {
	osc := pflag.StringP("exec", "e", "", "Execute the given command directly instead of interactive mode")
	pflag.Parse()

	app := App{handlers: HandlerMap{}}
	app.Register("state", FetchState)
	app.Register("appendLogs", AppendLogs)

	if *osc != "" {
		command := ParseCommand(*osc)
		result, err := app.Exec(command)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else {
			fmt.Println(result)
			return
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("command ? ")
		scanner.Scan()
		input := scanner.Text()
		command := ParseCommand(input)
		result, err := app.Exec(command)
		if err != nil {
			fmt.Printf("[ERROR] > %s\n", err)
		} else {
			fmt.Printf("[OK] > %s\n", result)
		}
	}
}
