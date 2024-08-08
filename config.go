package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// pinot-exporter can discover the Kubernetes services of Pinot Controller using a Label selector search
// You can configure the labels here or any Kubernetes discovery specific options.
type ServiceDiscoveryConfigK8S struct {
	Labels map[string]string `json:"labelSelector" yaml:"labelSelector"`
}
type Config struct {
	ListenPort            int              `json:"port" yaml:"port"`
	PinotController       *PinotController `json:"controller" yaml:"controller"`
	PollFrequencySeconds  int              `json:"poll_freq_seconds" yaml:"poll_freq_seconds"`
	MaxParallelCollectors int              `json:"max_parallel_collectors" yaml:"max_parallel_collectors"`
	// Mode can be [ "kubernetes", "direct"]
	Mode             string                    `json:"mode" yaml:"mode"`
	ServiceDiscovery ServiceDiscoveryConfigK8S `json:"serviceDiscovery" yaml:"serviceDiscovery"`
}

type Option func(*Config)

func NewConfig(options ...func(*Config)) *Config {

	// Start with some defaults where possible
	config := &Config{
		ListenPort:            8080,
		PollFrequencySeconds:  30,
		MaxParallelCollectors: 5,
		//PinotController:       &pinotDefault,
		Mode: "direct",
	}

	for _, opt := range options {
		opt(config)
	}
	return config
}

func (c *Config) IsValid() error {

	if (c.Mode != "direct") && (c.Mode != "kubernetes") {
		return fmt.Errorf("unknown mode %s - should be one of 'direct' or 'kubernetes'", c.Mode)
	}
	if c.Mode == "direct" {
		if c.PinotController == nil {
			return fmt.Errorf("Pinot controller config missing")
		}
	}
	if c.Mode == "kubernetes" {
		// First, make sure we have labels defined
		if len(c.ServiceDiscovery.Labels) == 0 {
			return fmt.Errorf("serviceDiscovery.labels is not defined")
		}
	}

	return nil
}

func WithMaxParallelCollectors(maxCollectors int) Option {
	return func(c *Config) {
		c.MaxParallelCollectors = maxCollectors
	}
}

// Listen on Port
func WithPort(port int) Option {
	return func(c *Config) {
		c.ListenPort = port
	}
}

// Pinot Controller URL + connectio ndetails
func WithPinotCluster(controller PinotController) Option {
	return func(c *Config) {
		c.PinotController = &controller
	}
}

// How often to poll target for metrics
func WithPollFrequencySeconds(frequencySeconds int) Option {
	return func(c *Config) {
		c.PollFrequencySeconds = frequencySeconds
	}
}

// Create a new Config from a YAML file
func NewConfigFromFile(filename string) (*Config, error) {
	config := NewConfig()
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return config, err
	}
	return config, nil
}
