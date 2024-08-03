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
}

func TestNewConfigWithOptions(t *testing.T) {

	config := NewConfig(
		WithPort(9999),
		WithPollFrequencySeconds(120),
	)
	// Assert we have some defaults as epected
	assert.Equal(t, 9999, config.ListenPort)
	assert.Equal(t, 120, config.PollFrequencySeconds)
}

func TestConfigIsValid(t *testing.T) {

	config := NewConfig()
	err := config.IsValid()
	// We are missing the Pinot controller info, which makes it invalid
	assert.NotNil(t, err)

	// Now assign a controller and test again
	controller := PinotController{
		URL: "http://localhost:9000",
	}
	config.PinotController = &controller
	err = config.IsValid()
	assert.Nil(t, err)

}
