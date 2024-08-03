package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type TableCache struct {
	Tables []string `json:"tables"`
	mutex  sync.Mutex
}

func refreshTableCache(ctx context.Context, controller *PinotController, sleepDuration int, tables chan<- []string) {
	// we never return.
	for {
		tableList, err := controller.ListTables(ctx)
		fmt.Printf("Discovered tables: %+v\n", tableList)
		if err != nil {
			panic(err)
		}
		tables <- tableList
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
}

// Refresh the table cache when we get a new list from the re
func (t *TableCache) TableRefreshChanListener(tables <-chan []string) {
	for {
		select {
		case newTables := <-tables:
			{
				fmt.Printf("TableCache received update: %+v\n", newTables)
				t.mutex.Lock()
				t.Tables = newTables
				t.mutex.Unlock()
			}
		default:
		}
	}
}

func (t *TableCache) GetTables() []string {
	return t.Tables
}
