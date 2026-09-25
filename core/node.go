package core

// Node represents a node in the network
type Node struct {
	ID NodeID

	// WARN: only for test, similiar to VRAM in placement policy
	TotalResource float64
	FreeResource  float64
	Cost          float64 // dollar every sec
}
