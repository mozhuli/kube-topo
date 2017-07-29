package types

// Topology is a representaion of a topological graph
type Topology struct {
	Name  string
	Links []*Link
}

// Link is a representation of a network link
type Link struct {
	Count   int32
	DstIP   string
	SrcIP   string
	DstPort string
	SrcPort string
}
