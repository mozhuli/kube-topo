package types

import (
	"sync"

	"github.com/mozhuli/kube-topo/pkg/sets"
)

// Topology is a representaion of a topological graph
type Topology struct {
	Name  string `json:"name"`
	Links []Link `json:"links"`
}

// IPLink specifies the link of podIP.
type IPLink struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// Link specifies the link of service.
type Link struct {
	DstSVC string `json:"dstServiceName,omitempty"`
	SrcSVC string `json:"srcServiceName,omitempty"`
	Count  int64  `json:"count"`
}

// TopoToIPs map topoName to the topo's all ips.
type TopoToIPs struct {
	core map[string]sets.String
	lock sync.RWMutex
}

func NewTopoToIPs() *TopoToIPs {
	var a TopoToIPs
	a.core = make(map[string]sets.String)
	return &a
}

func (a *TopoToIPs) Read(topoName string) (sets.String, bool) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	ips, ok := a.core[topoName]
	return ips, ok
}

func (a *TopoToIPs) Write(topoName string, ip string) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if ip == "" {
		a.core[topoName] = sets.String{}
	} else {
		a.core[topoName].Insert(ip)
	}
}
