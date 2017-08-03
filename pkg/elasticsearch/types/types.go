package types

// Topology is a representaion of a topological graph
type Topology struct {
	Name  string `json:"name"`
	Links []Link `json:"links"`
}

// Link specifies the link of pod.
type Link struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
	DstIP string `json:"dstIP,omitempty"`
	SrcIP string `json:"srcIP,omitempty"`
}
