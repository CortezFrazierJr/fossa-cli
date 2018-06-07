package v1

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"

	"github.com/fossas/fossa-cli/module"
	"github.com/fossas/fossa-cli/pkg"
)

var (
	ErrWrongVersion = errors.New("config file version is not 1")
)

type File struct {
	Version int `yaml:"version"`

	CLI     CLIProperties
	Analyze AnalyzeProperties

	// Internal computed + cached properties.
	modules []module.Module
}

type CLIProperties struct {
	// Upload configuration.
	APIKey   string `yaml:"api_key,omitempty"`
	Server   string `yaml:"server,omitempty"`
	Fetcher  string `yaml:"fetcher,omitempty"` // Defaults to custom
	Project  string `yaml:"project,omitempty"`
	Title    string `yaml:"title,omitempty"`
	Revision string `yaml:"revision,omitempty"`
	Branch   string `yaml:"branch,omitempty"` // Only used with custom fetcher
}

type AnalyzeProperties struct {
	Modules []ModuleProperties `yaml:"modules,omitempty"`
}

type ModuleProperties struct {
	Name    string                 `yaml:"name"`
	Path    string                 `yaml:"path"`
	Type    string                 `yaml:"type"`
	Options map[string]interface{} `yaml:"options,omitempty"`
}

func New(data []byte) (*File, error) {
	// Check whether version is correct. We first unmarshal into a map so that if
	// the type of `version` is not `int`, we can identify that issue distinct
	// from malformed YAML and handle it specially.
	var contents map[string]interface{}
	err := yaml.Unmarshal(data, &contents)
	if err != nil {
		return nil, err
	}
	if v, ok := contents["version"].(int); !ok || v != 1 {
		return nil, ErrWrongVersion
	}

	// Use mapstructure to fill out the File struct.
	var file File
	err = mapstructure.Decode(contents, &file)
	if err != nil {
		return nil, err
	}

	// Parse module configurations into modules.
	for _, config := range file.Analyze.Modules {
		// Parse and validate module type.
		t, err := pkg.ParseType(config.Type)
		if err != nil {
			return nil, errors.Wrapf(err, "could not parse module type %s", config.Type)
		}

		file.modules = append(file.modules, module.Module{
			Name:        config.Name,
			Type:        t,
			BuildTarget: config.Path,
			Dir:         config.Path,
			Options:     config.Options,
		})
	}

	return &file, nil
}

func (file *File) APIKey() string {
	return file.CLI.APIKey
}

func (file *File) Server() string {
	return file.CLI.Server
}

func (file *File) Title() string {
	if file.CLI.Title == "" {
		return file.CLI.Project
	}
	return file.CLI.Title
}

func (file *File) Fetcher() string {
	return file.CLI.Fetcher
}

func (file *File) Project() string {
	return file.CLI.Project
}

func (file *File) Branch() string {
	return file.CLI.Branch
}

func (file *File) Revision() string {
	return file.CLI.Revision
}

func (file *File) Modules() []module.Module {
	return file.modules
}
