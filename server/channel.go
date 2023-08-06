package server

import (
	"net/rpc"
	"time"
)

// Channel abstracts the communication channel between nodes
type Channel interface {
	Dest() Addr                                // Dest returns the destination address
	Call(method string, args, reply any) error // Call sends a RPC to the destination
}

// ReliableChannel is a basic channel without delay and loss
type ReliableChannel struct {
	dest Addr
}

// NewReliableChannel creates a ReliableChannel to the given destination
func NewReliableChannel(dest Addr) Channel {
	return &ReliableChannel{dest}
}

func (c *ReliableChannel) Dest() Addr {
	return c.dest
}

func (c *ReliableChannel) Call(method string, args, reply any) error {
	client, err := rpc.Dial("tcp", string(c.Dest()))
	if err != nil {
		return err
	}
	defer client.Close()

	return client.Call(method, args, reply)
}

// DelayedChannel is a channel with communication delay
type DelayedChannel struct {
	dest  Addr
	delay float64
}

// NewDelayedChannel creates a DelayedChannel with certain delay to the given destination
func NewDelayedChannel(dest Addr, delay float64) Channel {
	return &DelayedChannel{dest, delay}
}

func (c *DelayedChannel) Dest() Addr {
	return c.dest
}

func (c *DelayedChannel) Call(method string, args, reply any) error {
	if c.delay > 0 {
		time.Sleep(time.Duration(c.delay * float64(time.Millisecond)))
	}
	return NewReliableChannel(c.Dest()).Call(method, args, reply)
}

// LostChannel is a channel through which any messages will be lost
type LostChannel ReliableChannel

// NewLostChannel creates a LostChannel to the given destination
func NewLostChannel(dest Addr) Channel {
	return &LostChannel{dest}
}

func (c *LostChannel) Dest() Addr {
	return c.dest
}

func (c *LostChannel) Call(_ string, _, _ any) error {
	return nil // Lost channel never sends out any messages
}
