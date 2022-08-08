package dispatcher

import (
	"net/rpc"

	"sc.y3/peer"
)

type Client struct {
	client *rpc.Client
}

func FindDispatcher(addr string) (*Client, error) {
	d := new(Client)
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	d.client = client
	return d, nil
}

func (d *Client) GetConnectedPeers(w *peer.Worker) (map[string]string, error) {
	var reply RequestConnectReply
	err := d.client.Call("Dispatcher.RequestConnect", RequestConnectArgs{w.Name(), w.Addr()}, &reply)
	return reply.Peers, err
}

func (d *Client) GetAddr(name string) (string, error) {
	var reply RequestAddrReply
	err := d.client.Call("Dispatcher.RequestAddr", RequestAddrArgs{name}, &reply)
	return reply.Addr, err
}
