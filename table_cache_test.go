package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTableRefreshChanListener(t *testing.T) {
	tables := make(chan []string)

	var cache TableCache

	go cache.TableRefreshChanListener(tables)
	table_list1 := []string{"trololo", "skata"}
	tables <- table_list1
	// This refresh is not instant and goes through the channel etc. Wait 50m s before testing
	time.Sleep(time.Duration(50) * time.Millisecond)
	assert.Equal(t, cache.GetTables(), table_list1)

	table_list2 := []string{"trololo"}
	tables <- table_list2
	// This refresh is not instant and goes through the channel etc. Wait 50m s before testing
	time.Sleep(time.Duration(50) * time.Millisecond)
	assert.Equal(t, cache.GetTables(), table_list2)
}
