package server

import "fmt"

// Addr is alias to a string of address
type Addr string

// String converts Addr into string
func (a *Addr) String() string {
	return string(*a)
}

// NodeInfo
type NodeInfo struct {
	Name     string `json:"name"`
	Endpoint Addr   `json:"endpoint"`
}

// String converts NodeInfo into string
func (ni *NodeInfo) String() string {
	return fmt.Sprintf("%s@%s", ni.Name, ni.Endpoint.String())
}
