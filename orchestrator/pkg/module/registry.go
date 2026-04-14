package module

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Registry holds all discovered module descriptors.
type Registry struct {
	modules map[string]*Descriptor
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{modules: make(map[string]*Descriptor)}
}

// LoadDir scans a directory for *.yaml manifest files and registers each module.
func (r *Registry) LoadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("scanning module dir %s: %w", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		path := filepath.Join(dir, name)
		desc, err := LoadDescriptor(path)
		if err != nil {
			return fmt.Errorf("loading module manifest %s: %w", path, err)
		}
		if _, exists := r.modules[desc.Name]; exists {
			return fmt.Errorf("duplicate module name %q in %s", desc.Name, path)
		}
		r.modules[desc.Name] = desc
	}
	return nil
}

// Get returns a module descriptor by name.
func (r *Registry) Get(name string) (*Descriptor, bool) {
	desc, ok := r.modules[name]
	return desc, ok
}

// All returns all registered module descriptors.
func (r *Registry) All() []*Descriptor {
	result := make([]*Descriptor, 0, len(r.modules))
	for _, desc := range r.modules {
		result = append(result, desc)
	}
	return result
}

// Filter returns modules matching the given names. If names is empty, returns all.
func (r *Registry) Filter(names []string) []*Descriptor {
	if len(names) == 0 {
		return r.All()
	}
	var result []*Descriptor
	for _, name := range names {
		if desc, ok := r.modules[name]; ok {
			result = append(result, desc)
		}
	}
	return result
}

// SuitesForEnv returns all suites across all modules (optionally filtered) that support the given environment.
func (r *Registry) SuitesForEnv(env EnvironmentType, moduleFilter []string) []QualifiedSuite {
	modules := r.Filter(moduleFilter)
	var result []QualifiedSuite
	for _, mod := range modules {
		for i := range mod.Suites {
			suite := &mod.Suites[i]
			if suite.SupportsEnv(env) {
				result = append(result, QualifiedSuite{
					Module: mod.Name,
					Suite:  suite,
				})
			}
		}
	}
	return result
}

// QualifiedSuite pairs a module name with a suite reference.
type QualifiedSuite struct {
	Module string
	Suite  *Suite
}
