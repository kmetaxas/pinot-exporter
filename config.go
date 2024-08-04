package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ListenPort            int              `json:"port" yaml:"port"`
	PinotController       *PinotController `json:"controller" yaml:"controller"`
	PollFrequencySeconds  int              `json:"poll_freq_seconds" yaml:"poll_freq_seconds"`
	MaxParallelCollectors int              `json:"max_parallel_collectors" yaml:"max_parallel_collectors"`
}

type Option func(*Config)

func NewConfig(options ...func(*Config)) *Config {

	pinotDefault := PinotController{
		URL: "http://localhost:9000",
	}

	// Start with some defaults where possible
	config := &Config{
		ListenPort:            8080,
		PollFrequencySeconds:  30,
		MaxParallelCollectors: 5,
		PinotController:       &pinotDefault,
	}

	for _, opt := range options {
		opt(config)
	}
	return config
}

func (c *Config) IsValid() error {

	if c.PinotController == nil {
		return fmt.Errorf("Pinot controller config missing")
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
