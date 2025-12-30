// Package model defines data structures for representing architecture graphs.
package model

// Graph represents an architecture graph with components (nodes) and links (edges).
type Graph struct {
	Nodes []Node `yaml:"components"`
	Edges []Edge `yaml:"links"`
}

// Node represents a component in the architecture graph.
// Entity types: package, struct, interface, function, method, external.
type Node struct {
	ID     string `yaml:"id"`
	Title  string `yaml:"title"`
	Entity string `yaml:"entity"`
}

// Edge represents a link between components in the architecture graph.
// Type values: contains, calls, uses, embeds, import.
type Edge struct {
	From   string `yaml:"from"`
	To     string `yaml:"to"`
	Method string `yaml:"method,omitempty"`
	Type   string `yaml:"type,omitempty"`
}
