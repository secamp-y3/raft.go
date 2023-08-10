package server

import (
	"fmt"
	"sync"
)

const (
	invalidAddr = Addr("")
)

// Cluster manages members of the server cluster
type Cluster struct {
	members []NodeInfo
	mu      sync.Mutex
}

// NewCluster creates a new cluster
func NewCluster() *Cluster {
	return &Cluster{members: []NodeInfo{}}
}

// Members returns a copy of cluster members list
func (c *Cluster) Members() []NodeInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	members := make([]NodeInfo, len(c.members))
	copy(members, c.members)
	return members
}

// Append registeres a new member with the given name and endpoint address, unless the given node is not registered.
func (c *Cluster) Append(node NodeInfo) AppendResult {
	c.mu.Lock()
	defer c.mu.Unlock()
	addr, exist := func() (Addr, bool) {
		for _, m := range c.members {
			if m.Name == node.Name {
				return m.Endpoint, true
			}
		}
		return invalidAddr, false
	}()
	if !exist {
		c.members = append(c.members, node)
		return append_result_success
	}
	if addr != node.Endpoint {
		return append_result_inconsistent
	}
	return append_result_exist
}

// Remove deletes the member with given name from the cluster.
//
// If the given name is not registered, this function only returns `false` without any change.
// Otherwise, this function deletes the member from the list and returns `true`.
func (c *Cluster) Remove(name string) (Addr, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	idx := func() int {
		for i, m := range c.members {
			if m.Name == name {
				return i
			}
		}
		return -1
	}()
	if idx >= 0 {
		addr := c.members[idx].Endpoint
		c.members[idx] = c.members[len(c.members)-1]
		c.members = c.members[:len(c.members)-1]
		return addr, true
	}
	return invalidAddr, false
}

// String describes the cluster in string format
func (c *Cluster) String() string {
	return fmt.Sprintf("Cluster{members: %d}", len(c.members))
}

// AppendResult represents a result of Cluster#Append
type AppendResult int

const (
	append_result_success AppendResult = iota
	append_result_exist
	append_result_inconsistent
)

// Success returns true when Cluster#Append successufully appends the given node.
func (r AppendResult) Success() bool {
	return r == append_result_success
}

// NotChange return true when Cluster#Append does not change the internal state.
func (r AppendResult) NotChanged() bool {
	return r == append_result_exist || r == append_result_inconsistent
}

// Inconsistent returns true when the given node name exists but the endpoint is not consistent with the registered one.
func (r AppendResult) Inconsistent() bool {
	return r == append_result_inconsistent
}
