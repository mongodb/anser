package anser

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/mongodb/grip"
	"github.com/tychoish/tarjan"
)

// DependencyNetworker provides answers to questions about the
// dependencies of a task and is available white generating
// migrations. Methods should do nothing in particular
//
// Implementations should be mutable and thread-safe.
type DependencyNetworker interface {
	// Add inserts a list of dependencies for a given item. If the
	// slice of dependencies is empty, Add is a noop. Furthermore,
	// the Add method provides no validation, and will do nothing
	// to prevent cycles or broken dependencies.
	Add(string, []string)

	// Resolve, returns all of the dependencies for the specified task.
	Resolve(string) []string

	// All returns a list of all tasks that have registered
	// dependencies.
	All() []string

	// Network returns the dependency graph for all registered
	// tasks as a mapping of task IDs to the IDs of its
	// dependencies.
	Network() map[string][]string

	// Validate returns errors if there are either dependencies
	// specified that do not have tasks available *or* if there
	// are dependency cycles.
	Validate() error

	// AddGroup and GetGroup set and return the lists of tasks
	// that beloing to a specific task group. Unlike the specific
	// task dependency setters.
	AddGroup(string, []string)
	GetGroup(string) []string

	// For introspection and convince, DependencyNetworker
	// composes implementations of common interfaces.
	fmt.Stringer
	json.Marshaler
}

func NewDependencyNetwork() DependencyNetworker {
	return &dependencyNetwork{
		network: make(map[string]map[string]struct{}),
		group:   make(map[string]map[string]struct{}),
	}
}

type dependencyNetwork struct {
	network map[string]map[string]struct{}
	group   map[string]map[string]struct{}
	mu      sync.RWMutex
}

func (n *dependencyNetwork) Add(name string, deps []string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	depSet, ok := n.network[name]
	if !ok {
		depSet = make(map[string]struct{})
		n.network[name] = depSet
	}

	for _, d := range deps {
		depSet[d] = struct{}{}
	}
}

func (n *dependencyNetwork) Resolve(name string) []string {
	out := []string{}

	n.mu.RLock()
	defer n.mu.RUnlock()

	edges, ok := n.network[name]
	if !ok {
		return out
	}

	for e := range edges {
		out = append(out, e)
	}

	return out
}

func (n *dependencyNetwork) All() []string {
	out := []string{}

	n.mu.RLock()
	defer n.mu.RUnlock()
	for name := range n.network {
		out = append(out, name)
	}

	return out
}

func (n *dependencyNetwork) Network() map[string][]string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.getNetworkUnsafe()
}

// this is implemented separately from network so we can use it in
// validation and have a sane locking strategy.
func (n *dependencyNetwork) getNetworkUnsafe() map[string][]string {
	out := make(map[string][]string)

	for node, edges := range n.network {
		deps := []string{}
		for e := range edges {
			deps = append(deps, e)
		}
		out[node] = deps
	}
	return out
}

func (n *dependencyNetwork) Validate() error {
	dependencies := make(map[string]struct{})
	catcher := grip.NewCatcher()

	n.mu.RLock()
	defer n.mu.RUnlock()

	graph := n.getNetworkUnsafe()
	for _, edges := range graph {
		for _, id := range edges {
			dependencies[id] = struct{}{}
		}
	}

	for id := range dependencies {
		if _, ok := n.network[id]; !ok {
			catcher.Add(fmt.Errorf("dependency %s is not defined", id))
		}
	}

	for _, group := range tarjan.Connections(graph) {
		if len(group) > 1 {
			catcher.Add(fmt.Errorf("cycle detected between nodes: [%s]",
				strings.Join(group, ", ")))
		}
	}

	return catcher.Resolve()
}

func (n *dependencyNetwork) AddGroup(name string, group []string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	groupSet, ok := n.group[name]
	if !ok {
		groupSet = make(map[string]struct{})
		n.group[name] = groupSet
	}

	for _, g := range group {
		groupSet[g] = struct{}{}
	}
}

func (n *dependencyNetwork) GetGroup(name string) []string {
	out := []string{}

	n.mu.RLock()
	defer n.mu.RUnlock()

	group, ok := n.group[name]
	if !ok {
		return out
	}

	for g := range group {
		out = append(out, g)
	}

	return out
}

//////////////////////////////////////////
//
// Output Formats

func (n *dependencyNetwork) MarshalJSON() ([]byte, error) { return json.Marshal(n.Network()) }
func (n *dependencyNetwork) String() string               { return fmt.Sprintf("%v", n.Network()) }
