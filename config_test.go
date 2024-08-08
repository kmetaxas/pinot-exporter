package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigFromFile(t *testing.T) {

	config, err := NewConfigFromFile("testdata/files/config.sample.yaml")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 8088, config.ListenPort, "Listen port should be 8088")

}

func TestNewConfig(t *testing.T) {

	config := NewConfig()
	// Assert we have some defaults as epected
	assert.Equal(t, 8080, config.ListenPort)
	assert.Equal(t, 30, config.PollFrequencySeconds)
	assert.Equal(t, 5, config.MaxParallelCollectors)
}

func TestNewConfigWithOptions(t *testing.T) {

	config := NewConfig(
		WithPort(9999),
		WithPollFrequencySeconds(120),
		WithMaxParallelCollectors(12),
	)
	// Assert we have some defaults as epected
	assert.Equal(t, 9999, config.ListenPort)
	assert.Equal(t, 120, config.PollFrequencySeconds)
	assert.Equal(t, 12, config.MaxParallelCollectors)
}

func TestConfigIsValid(t *testing.T) {

	config := NewConfig()
	assert.NotNil(t, config.IsValid())

	controller := PinotController{
		URL: "http://localhost:9000",
	}
	// Now remove a controller and test again
	config.PinotController = &controller

	assert.Nil(t, config.IsValid())

	// switch to kubernetes mode
	config.Mode = "kubernetes"
	// validation should fail as we have no Labels
	assert.NotNil(t, config.IsValid())

	// add labels and make sure validation passes
	kubeconfig := ServiceDiscoveryConfigK8S{
		Labels: map[string]string{
			"app":      "pinot",
			"nodeType": "controller",
		},
	}
	config.ServiceDiscovery = kubeconfig
	assert.Nil(t, config.IsValid())

}
