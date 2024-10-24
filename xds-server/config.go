package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultScanInterval = 10
const DefaultName = "xds-server"

type Service struct {
	Name        string   `yaml:"name"`
	FullName    string   `yaml:"full_name"`
	Zone        string   `yaml:"zone"`
	DefaultPort int      `yaml:"default_port"`
	Endpoints   []string `yaml:"endpoints"`
}

func (s *Service) updateEndpoints() {
	default_port := s.DefaultPort
	if default_port == 0 {
		return
	}
	for i := range s.Endpoints {
		endpoint := s.Endpoints[i]
		splits := strings.SplitN(endpoint, ":", 2)
		if len(splits) == 1 {
			s.Endpoints[i] = fmt.Sprintf("%s:%d", endpoint, default_port)
		}
	}
}

type Config struct {
	Name         string    `yaml:"name"`
	ScanInterval int       `yaml:"scan_interval"`
	Services     []Service `yaml:"services"`
}

func ReadConfig(cfgFile string) (Config, error) {
	cfg := Config{}

	if cfgFile == "" {
		return cfg, fmt.Errorf("empty config file ")
	}

	if _, err := os.Stat(cfgFile); errors.Is(err, os.ErrNotExist) {
		return cfg, fmt.Errorf("Config file : %s , does not exist", cfgFile)
	}

	yaml_file, err := os.ReadFile(cfgFile)
	if err != nil {
		return cfg, fmt.Errorf("unable to read config file : %s", cfgFile)
	}

	err = yaml.Unmarshal(yaml_file, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("invalid yaml file : %s , Error %w", cfgFile, err)
	}

	if cfg.Name == "" {
		cfg.Name = DefaultName
	}

	if cfg.ScanInterval == 0 {
		cfg.ScanInterval = DefaultScanInterval
	}

	for i := range cfg.Services {
		cfg.Services[i].updateEndpoints()
	}
	return cfg, nil
}
