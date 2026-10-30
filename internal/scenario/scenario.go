// Package scenario defines the versioned JSON file format used by the CLI,
// tests, and the web trace viewer. Scenario files are untrusted input: they
// are size-limited, strictly decoded, and validated before execution.
package scenario

// SchemaVersion is the scenario format understood by this build.
const SchemaVersion = "1.0"

// Scenario is a complete, reproducible experiment description.
type Scenario struct {
	SchemaVersion string   `json:"schemaVersion"`
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	Seed          uint64   `json:"seed"`
	Nodes         []Node   `json:"nodes"`
	Links         []Link   `json:"links"`
	Actions       []Action `json:"actions,omitempty"`
}

// Node is a host or router in the scenario file.
type Node struct {
	ID         string      `json:"id"`
	Kind       string      `json:"kind"`
	Name       string      `json:"name,omitempty"`
	Down       bool        `json:"down,omitempty"`
	Interfaces []Interface `json:"interfaces"`
	Routes     []Route     `json:"routes,omitempty"`
}

// Interface is an addressed attachment point.
type Interface struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

// Route is one static route table entry.
type Route struct {
	Network   string `json:"network"`
	NextHop   string `json:"nextHop,omitempty"`
	Interface string `json:"interface"`
	Metric    int    `json:"metric,omitempty"`
}

// Link is a point-to-point connection.
type Link struct {
	ID           string `json:"id"`
	A            string `json:"a"`
	B            string `json:"b"`
	DelayMs      int64  `json:"delayMs"`
	BandwidthBps uint64 `json:"bandwidthBps,omitempty"`
	QueueLimit   int    `json:"queueLimit,omitempty"`
	Down         bool   `json:"down,omitempty"`
}

// Action is a scheduled experiment step.
type Action struct {
	AtMs       int64  `json:"atMs"`
	Kind       string `json:"kind"`
	From       string `json:"from"`
	To         string `json:"to"`
	Count      int    `json:"count,omitempty"`
	IntervalMs int64  `json:"intervalMs,omitempty"`
	TTL        uint8  `json:"ttl,omitempty"`
}
