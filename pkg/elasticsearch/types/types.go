package types

// Topology is a representaion of a topological graph
type Topology struct {
	Name  string    `json:"name"`
	Links []LinkSVC `json:"links"`
}

// LinkSVC specifies the link of SVC.
type LinkSVC struct {
	Key    string `json:"key"`
	SrcSVC string `json:"srcSVC"`
	DstSVC string `json:"dstSVC"`
	Count  int64  `json:"count"`
}

// Link specifies the link of pod.
type Link struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
	DstIP string `json:"dstIP,omitempty"`
	SrcIP string `json:"srcIP,omitempty"`
}
