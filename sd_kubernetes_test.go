package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLabelSelectorString(t *testing.T) {
	labels := map[string]string{
		"skata":    "pola",
		"trolling": "maximum",
	}
	labelString := GetLabelSelectorString(labels)
	assert.Equal(t, "skata=pola,trolling=maximum", labelString)

}
