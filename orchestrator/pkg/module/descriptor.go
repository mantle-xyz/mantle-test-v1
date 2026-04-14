package module

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Descriptor represents a module's mantle-test.yaml manifest.
type Descriptor struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Source      SourceConfig    `yaml:"source"`
	Suites      []Suite         `yaml:"suites"`
}

// SourceConfig defines where to get and how to run the module.
type SourceConfig struct {
	Local *LocalSource `yaml:"local,omitempty"`
	CI    *CISource    `yaml:"ci,omitempty"`
}

// LocalSource: clone repo and run command locally.
type LocalSource struct {
	Repo   string `yaml:"repo"`
	Branch string `yaml:"branch"`
	Path   string `yaml:"path"`
}

// CISource: trigger target repo's CI workflow.
type CISource struct {
	Repo     string `yaml:"repo"`     // "owner/repo"
	Workflow string `yaml:"workflow"` // "mantle-test.yaml"
	Event    string `yaml:"event"`    // "mantle-test"
}

// Suite defines a single test suite within a module.
type Suite struct {
	Name         string            `yaml:"name"`
	Phase        Phase             `yaml:"phase"`
	Environments []EnvironmentType `yaml:"environments"`
	Command      string            `yaml:"command"`
	EnvVars      []string          `yaml:"env_vars,omitempty"`
	ResultFormat ResultFormat      `yaml:"result_format"`
	ResultPath   string            `yaml:"result_path,omitempty"`
	Timeout      Duration          `yaml:"timeout"`
	DependsOn    []string          `yaml:"depends_on,omitempty"`
}



// SupportsEnv returns true if the suite can run in the given environment.
func (s *Suite) SupportsEnv(env EnvironmentType) bool {
	for _, e := range s.Environments {
		if e == env {
			return true
		}
	}
	return false
}

// QualifiedName returns "module:suite" identifier.
func (s *Suite) QualifiedName(moduleName string) string {
	return moduleName + ":" + s.Name
}

// Duration wraps time.Duration for YAML unmarshaling of strings like "30m", "2h".
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = parsed
	return nil
}

func (d Duration) MarshalYAML() (interface{}, error) {
	return d.Duration.String(), nil
}

// LoadDescriptor reads and parses a mantle-test.yaml manifest file.
func LoadDescriptor(path string) (*Descriptor, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest %s: %w", path, err)
	}
	var desc Descriptor
	if err := yaml.Unmarshal(data, &desc); err != nil {
		return nil, fmt.Errorf("parsing manifest %s: %w", path, err)
	}
	if desc.Name == "" {
		return nil, fmt.Errorf("manifest %s: name is required", path)
	}
	return &desc, nil
}
